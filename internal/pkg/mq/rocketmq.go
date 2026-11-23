package mq

import (
	"context"
	"fmt"
	"strings"

	rmq "github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"

	"redbook-go/internal/config"
)

// RocketMQProducer wraps a RocketMQ sync producer. When sharded is enabled it
// uses a hash queue selector so messages with the same sharding key land in the
// same queue.
type RocketMQProducer struct {
	producer rmq.Producer
}

func NewRocketMQProducer(cfg config.RocketMQConfig, groupSuffix string, sharded bool) (*RocketMQProducer, error) {
	opts := []producer.Option{
		producer.WithNameServer(parseNameServers(cfg.NameServer)),
		producer.WithGroupName(groupName(cfg.ProducerGroup, groupSuffix)),
		producer.WithRetry(2),
	}
	if sharded {
		opts = append(opts, producer.WithQueueSelector(producer.NewHashQueueSelector()))
	}
	p, err := rmq.NewProducer(opts...)
	if err != nil {
		return nil, fmt.Errorf("new rocketmq producer: %w", err)
	}
	if err := p.Start(); err != nil {
		return nil, fmt.Errorf("start rocketmq producer: %w", err)
	}
	return &RocketMQProducer{producer: p}, nil
}

func (p *RocketMQProducer) Publish(ctx context.Context, topic string, payload []byte) error {
	if p == nil || p.producer == nil {
		return fmt.Errorf("rocketmq producer is not initialized")
	}
	_, err := p.producer.SendSync(ctx, primitive.NewMessage(topic, payload))
	return err
}

func (p *RocketMQProducer) PublishSharded(ctx context.Context, topic string, payload []byte, shardingKey string) error {
	if p == nil || p.producer == nil {
		return fmt.Errorf("rocketmq producer is not initialized")
	}
	msg := primitive.NewMessage(topic, payload)
	msg.WithShardingKey(shardingKey)
	_, err := p.producer.SendSync(ctx, msg)
	return err
}

func (p *RocketMQProducer) Close() error {
	if p == nil || p.producer == nil {
		return nil
	}
	return p.producer.Shutdown()
}

// RocketMQTransactionProducer sends transactional messages. The listener runs
// the local transaction, and the broker re-checks with the listener when needed.
type RocketMQTransactionProducer struct {
	producer rmq.TransactionProducer
}

func NewRocketMQTransactionProducer(cfg config.RocketMQConfig, groupSuffix string, listener primitive.TransactionListener) (*RocketMQTransactionProducer, error) {
	p, err := rmq.NewTransactionProducer(listener,
		producer.WithNameServer(parseNameServers(cfg.NameServer)),
		producer.WithGroupName(groupName(cfg.ProducerGroup, groupSuffix)),
		producer.WithRetry(2),
	)
	if err != nil {
		return nil, fmt.Errorf("new rocketmq transaction producer: %w", err)
	}
	if err := p.Start(); err != nil {
		return nil, fmt.Errorf("start rocketmq transaction producer: %w", err)
	}
	return &RocketMQTransactionProducer{producer: p}, nil
}

func (p *RocketMQTransactionProducer) PublishInTransaction(ctx context.Context, topic string, payload []byte) error {
	if p == nil || p.producer == nil {
		return fmt.Errorf("rocketmq transaction producer is not initialized")
	}
	res, err := p.producer.SendMessageInTransaction(ctx, primitive.NewMessage(topic, payload))
	if err != nil {
		return err
	}
	if res.State == primitive.RollbackMessageState {
		return fmt.Errorf("transaction rolled back")
	}
	return nil
}

func (p *RocketMQTransactionProducer) Close() error {
	if p == nil || p.producer == nil {
		return nil
	}
	return p.producer.Shutdown()
}

// MessageHandler processes a single decoded message payload. Returning an error
// causes the consumer to retry the message.
type MessageHandler func(ctx context.Context, payload []byte) error

type RocketMQConsumer struct {
	consumer rmq.PushConsumer
}

func NewRocketMQConsumer(cfg config.RocketMQConfig, groupSuffix, topic string, orderly bool, handler MessageHandler) (*RocketMQConsumer, error) {
	c, err := rmq.NewPushConsumer(
		consumer.WithNameServer(parseNameServers(cfg.NameServer)),
		consumer.WithGroupName(groupName(cfg.ConsumerGroup, groupSuffix)),
		consumer.WithConsumerOrder(orderly),
	)
	if err != nil {
		return nil, fmt.Errorf("new rocketmq consumer: %w", err)
	}
	if err := c.Subscribe(topic, consumer.MessageSelector{}, func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for _, msg := range msgs {
			if err := handler(ctx, msg.Body); err != nil {
				return consumer.ConsumeRetryLater, nil
			}
		}
		return consumer.ConsumeSuccess, nil
	}); err != nil {
		return nil, fmt.Errorf("subscribe topic %s: %w", topic, err)
	}
	if err := c.Start(); err != nil {
		return nil, fmt.Errorf("start rocketmq consumer: %w", err)
	}
	return &RocketMQConsumer{consumer: c}, nil
}

func (c *RocketMQConsumer) Close() error {
	if c == nil || c.consumer == nil {
		return nil
	}
	return c.consumer.Shutdown()
}

func groupName(base, suffix string) string {
	if suffix == "" {
		return base
	}
	return base + "-" + suffix
}

func parseNameServers(s string) primitive.NamesrvAddr {
	var out primitive.NamesrvAddr
	for _, part := range strings.Split(s, ",") {
		if v := strings.TrimSpace(part); v != "" {
			out = append(out, v)
		}
	}
	return out
}
