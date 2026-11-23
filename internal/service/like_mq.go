package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"redbook-go/internal/config"
	"redbook-go/internal/model"
	"redbook-go/internal/pkg/idgen"
	"redbook-go/internal/pkg/mq"
)

// LikeMQ mirrors the original Java like pipeline:
//
//	click-topic --dedupe + transactional message--> click-record-topic
//	click-record-topic --Redis Lua aggregation--> click-summary-topic
//	click-summary-topic --persist--> click_summary
type LikeMQ struct {
	cfg   *config.Config
	db    *gorm.DB
	redis *redis.Client

	clickProducer   *mq.RocketMQProducer
	recordProducer  *mq.RocketMQTransactionProducer
	summaryProducer *mq.RocketMQProducer

	clickConsumer   *mq.RocketMQConsumer
	recordConsumer  *mq.RocketMQConsumer
	summaryConsumer *mq.RocketMQConsumer

	localCache sync.Map
}

type ClickEvent struct {
	NoteID    int64 `json:"noteId"`
	UserID    int64 `json:"userId"`
	ClickType int32 `json:"clickType"`
}

func NewLikeMQ(cfg *config.Config, db *gorm.DB, rdb *redis.Client) (*LikeMQ, error) {
	m := &LikeMQ{cfg: cfg, db: db, redis: rdb}

	var err error
	m.clickProducer, err = mq.NewRocketMQProducer(cfg.RocketMQ, "ClickProducer", true)
	if err != nil {
		return nil, err
	}
	m.summaryProducer, err = mq.NewRocketMQProducer(cfg.RocketMQ, "ClickSummaryProducer", true)
	if err != nil {
		m.close()
		return nil, err
	}
	m.recordProducer, err = mq.NewRocketMQTransactionProducer(cfg.RocketMQ, "ClickRecordProducer", &likeTransactionListener{m: m})
	if err != nil {
		m.close()
		return nil, err
	}

	m.clickConsumer, err = mq.NewRocketMQConsumer(cfg.RocketMQ, "ClickConsumer", cfg.RocketMQ.ClickTopic, true, m.handleClick)
	if err != nil {
		m.close()
		return nil, err
	}
	m.recordConsumer, err = mq.NewRocketMQConsumer(cfg.RocketMQ, "ClickRecordConsumer", cfg.RocketMQ.ClickRecordTopic, false, m.handleClickRecord)
	if err != nil {
		m.close()
		return nil, err
	}
	m.summaryConsumer, err = mq.NewRocketMQConsumer(cfg.RocketMQ, "ClickSummaryConsumer", cfg.RocketMQ.ClickSummaryTopic, true, m.handleClickSummary)
	if err != nil {
		m.close()
		return nil, err
	}
	return m, nil
}

func (m *LikeMQ) Stop() {
	m.close()
}

func (m *LikeMQ) close() {
	if m.clickConsumer != nil {
		_ = m.clickConsumer.Close()
	}
	if m.recordConsumer != nil {
		_ = m.recordConsumer.Close()
	}
	if m.summaryConsumer != nil {
		_ = m.summaryConsumer.Close()
	}
	if m.recordProducer != nil {
		_ = m.recordProducer.Close()
	}
	if m.summaryProducer != nil {
		_ = m.summaryProducer.Close()
	}
	if m.clickProducer != nil {
		_ = m.clickProducer.Close()
	}
}

func (m *LikeMQ) PublishClick(userID, noteID int64, clickType int32) (ClickResult, error) {
	payload, err := json.Marshal(ClickEvent{NoteID: noteID, UserID: userID, ClickType: clickType})
	if err != nil {
		return ClickResult{Clicked: false}, err
	}
	if err := m.clickProducer.PublishSharded(context.Background(), m.cfg.RocketMQ.ClickTopic, payload, strconv.FormatInt(userID, 10)); err != nil {
		return ClickResult{Clicked: false}, err
	}
	return ClickResult{Clicked: true}, nil
}

// handleClick consumes click-topic: dedupe, build a ClickRecords, then send it
// to click-record-topic as a transactional message.
func (m *LikeMQ) handleClick(ctx context.Context, payload []byte) error {
	var ev ClickEvent
	if err := json.Unmarshal(payload, &ev); err != nil {
		return err
	}
	noteID, userID, clickType := ev.NoteID, ev.UserID, ev.ClickType

	cacheKey := strconv.FormatInt(userID, 10) + "#" + strconv.FormatInt(noteID, 10)
	if v, ok := m.localCache.Load(cacheKey); ok && v.(int32) == clickType {
		return nil
	}

	if m.redis != nil {
		if ok, _ := m.redis.SIsMember(ctx, likeSetKey(userID), likeMember(noteID, clickType)).Result(); ok {
			m.localCache.Store(cacheKey, clickType)
			return nil
		}
	}

	var record model.ClickRecords
	err := m.db.WithContext(ctx).Where("receiver_id = ? AND clicker_id = ?", noteID, userID).
		Order("created DESC").First(&record).Error
	if err == nil {
		m.updateLikeSet(ctx, userID, noteID, record.ClickType)
		m.localCache.Store(cacheKey, record.ClickType)
		if record.ClickType == clickType {
			return nil
		}
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		if clickType == -1 {
			m.updateLikeSet(ctx, userID, noteID, -1)
			m.localCache.Store(cacheKey, int32(-1))
			return nil
		}
	} else {
		return err
	}

	record = model.ClickRecords{
		ID:         idgen.NextID(),
		ReceiverID: noteID,
		ClickerID:  userID,
		ClickType:  clickType,
		Created:    time.Now(),
	}
	body, err := json.Marshal(record)
	if err != nil {
		return err
	}
	return m.recordProducer.PublishInTransaction(ctx, m.cfg.RocketMQ.ClickRecordTopic, body)
}

// handleClickRecord consumes click-record-topic: aggregates increments per note
// with a Redis Lua script, then forwards the lock-holder record to click-summary-topic.
func (m *LikeMQ) handleClickRecord(ctx context.Context, payload []byte) error {
	var record model.ClickRecords
	if err := json.Unmarshal(payload, &record); err != nil {
		return err
	}
	if m.redis == nil {
		return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			_, err := applyClickSummary(tx, record.ReceiverID, int64(record.ClickType))
			return err
		})
	}

	receiverID := strconv.FormatInt(record.ReceiverID, 10)
	recordID := strconv.FormatInt(record.ID, 10)
	lockKey := fmt.Sprintf("lock#{%d}", record.ReceiverID)
	cntRecordKey := fmt.Sprintf("cnt#{%d}#%d", record.ReceiverID, record.ID)

	script := `local receiverId = ARGV[1]
local recordId = ARGV[2]
local lockKey = KEYS[1]
local cntRecordKey = KEYS[2]
local setnx_result = redis.call("SETNX", lockKey, recordId)
if setnx_result == 1 then
    redis.call("INCRBY", cntRecordKey, tonumber(ARGV[3]))
    return recordId
else
    local idToUse = redis.call("GET", lockKey)
    if idToUse == recordId then
        return recordId
    end
    local cntKeyForIdToUse = "cnt#{" .. receiverId .. "}#" .. idToUse
    redis.call("INCRBY", cntKeyForIdToUse, tonumber(ARGV[3]))
    return idToUse
end`

	res, err := m.redis.Eval(ctx, script, []string{lockKey, cntRecordKey}, receiverID, recordID, strconv.FormatInt(int64(record.ClickType), 10)).Result()
	if err != nil {
		return err
	}
	if used, ok := res.(string); ok && used == recordID {
		return m.summaryProducer.PublishSharded(ctx, m.cfg.RocketMQ.ClickSummaryTopic, payload, strconv.FormatInt(record.ReceiverID, 10))
	}
	return nil
}

// handleClickSummary consumes click-summary-topic: releases the lock, reads the
// aggregated increment and persists it into click_summary.
func (m *LikeMQ) handleClickSummary(ctx context.Context, payload []byte) error {
	var record model.ClickRecords
	if err := json.Unmarshal(payload, &record); err != nil {
		return err
	}
	if m.redis == nil {
		return nil
	}

	recordID := strconv.FormatInt(record.ID, 10)
	lockKey := fmt.Sprintf("lock#{%d}", record.ReceiverID)
	cntRecordKey := fmt.Sprintf("cnt#{%d}#%d", record.ReceiverID, record.ID)
	cntKey := fmt.Sprintf("cnt#{%d}", record.ReceiverID)

	release := `local key = KEYS[1]
local idToUse = redis.call("GET", key)
if ARGV[1] == idToUse then
    redis.call("DEL", key)
end`
	if _, err := m.redis.Eval(ctx, release, []string{lockKey}, recordID).Result(); err != nil {
		return err
	}

	cntStr, err := m.redis.Get(ctx, cntRecordKey).Result()
	if err != nil || cntStr == "" {
		return nil
	}
	cnt, err := strconv.ParseInt(cntStr, 10, 64)
	if err != nil {
		return err
	}

	var newCount int64
	err = m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var e error
		newCount, e = applyClickSummary(tx, record.ReceiverID, cnt)
		return e
	})
	if err != nil {
		return err
	}
	m.redis.Set(ctx, cntKey, strconv.FormatInt(newCount, 10), 3*24*time.Hour)
	m.redis.Del(ctx, cntRecordKey)
	return nil
}

func likeSetKey(clickerID int64) string {
	return fmt.Sprintf("like#records#{%d}#%s", clickerID, time.Now().Format("20060102"))
}

func likeMember(receiverID int64, clickType int32) string {
	if clickType == 1 {
		return "+" + strconv.FormatInt(receiverID, 10)
	}
	return "-" + strconv.FormatInt(receiverID, 10)
}

func (m *LikeMQ) updateLikeSet(ctx context.Context, clickerID, receiverID int64, clickType int32) {
	if m.redis == nil {
		return
	}
	setKey := likeSetKey(clickerID)
	use := likeMember(receiverID, clickType)
	notUse := likeMember(receiverID, -clickType)
	script := `local set_key = KEYS[1]
redis.call('SREM', set_key, ARGV[2])
redis.call('SADD', set_key, ARGV[1])
local set_size = redis.call('SCARD', set_key)
if set_size == 1 then
    redis.call('EXPIRE', set_key, 90000)
end
if set_size > tonumber(ARGV[3]) then
    local to_remove = math.floor(set_size / 2)
    local elements = redis.call('SMEMBERS', set_key)
    for i = 1, to_remove do
        redis.call('SREM', set_key, elements[i])
    end
end`
	_ = m.redis.Eval(ctx, script, []string{setKey}, use, notUse, "100").Err()
}

type likeTransactionListener struct {
	m *LikeMQ
}

func (l *likeTransactionListener) ExecuteLocalTransaction(msg *primitive.Message) primitive.LocalTransactionState {
	var record model.ClickRecords
	if err := json.Unmarshal(msg.Body, &record); err != nil {
		return primitive.RollbackMessageState
	}
	if err := l.m.db.Create(&record).Error; err != nil {
		return primitive.UnknowState
	}
	cacheKey := strconv.FormatInt(record.ClickerID, 10) + "#" + strconv.FormatInt(record.ReceiverID, 10)
	l.m.localCache.Store(cacheKey, record.ClickType)
	l.m.updateLikeSet(context.Background(), record.ClickerID, record.ReceiverID, record.ClickType)
	return primitive.CommitMessageState
}

func (l *likeTransactionListener) CheckLocalTransaction(msg *primitive.MessageExt) primitive.LocalTransactionState {
	var record model.ClickRecords
	if err := json.Unmarshal(msg.Body, &record); err != nil {
		return primitive.RollbackMessageState
	}
	var found model.ClickRecords
	if err := l.m.db.First(&found, record.ID).Error; err == nil {
		return primitive.CommitMessageState
	}
	if record.Created.Add(20 * time.Second).After(time.Now()) {
		return primitive.UnknowState
	}
	return primitive.RollbackMessageState
}
