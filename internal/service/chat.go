package service

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"redbook-go/internal/model"
)

func (s *Service) ChatEnsureThread(userID, peerID int64) (int64, error) {
	if userID == peerID {
		return 0, errors.New("不能与自己会话")
	}
	low, high := userID, peerID
	if low > high {
		low, high = high, low
	}
	var thread model.ChatThread
	err := s.DB.Where("user_low = ? AND user_high = ?", low, high).First(&thread).Error
	if err == nil {
		return thread.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}
	thread = model.ChatThread{UserLow: low, UserHigh: high}
	if err := s.DB.Create(&thread).Error; err != nil {
		return 0, err
	}
	return thread.ID, nil
}

func (s *Service) ChatSend(userID, toUserID int64, content string) (int64, error) {
	if content == "" {
		return 0, errors.New("消息不能为空")
	}
	threadID, err := s.ChatEnsureThread(userID, toUserID)
	if err != nil {
		return 0, err
	}
	now := time.Now()
	msg := model.ChatMessage{ThreadID: threadID, SenderID: userID, Content: content, CreateTime: now}
	if err := s.DB.Create(&msg).Error; err != nil {
		return 0, err
	}
	s.DB.Model(&model.ChatThread{}).Where("id = ?", threadID).Update("last_msg_time", now)
	return msg.ID, nil
}

func (s *Service) ChatThreads(userID int64, current, size int64) ([]map[string]any, int64, error) {
	var threads []model.ChatThread
	var total int64
	s.DB.Model(&model.ChatThread{}).Where("user_low = ? OR user_high = ?", userID, userID).Count(&total)
	if err := s.DB.Where("user_low = ? OR user_high = ?", userID, userID).
		Order("last_msg_time DESC").Offset(int((current - 1) * size)).Limit(int(size)).Find(&threads).Error; err != nil {
		return nil, 0, err
	}
	list := make([]map[string]any, 0, len(threads))
	for _, t := range threads {
		peerID := t.UserLow
		if peerID == userID {
			peerID = t.UserHigh
		}
		var peer model.User
		s.DB.First(&peer, peerID)
		list = append(list, map[string]any{
			"id":          t.ID,
			"userLow":     t.UserLow,
			"userHigh":    t.UserHigh,
			"lastMsgTime": t.LastMsgTime,
			"peerNickName": peer.NickName,
			"peerIcon":    peer.Icon,
			"unreadCount": s.privateUnread(t, userID),
		})
	}
	return list, total, nil
}

func (s *Service) ChatMessages(userID, threadID, current, size int64) ([]model.ChatMessage, int64, error) {
	var thread model.ChatThread
	if err := s.DB.First(&thread, threadID).Error; err != nil {
		return nil, 0, errors.New("会话不存在")
	}
	if thread.UserLow != userID && thread.UserHigh != userID {
		return nil, 0, errors.New("无权访问该会话")
	}
	var msgs []model.ChatMessage
	var total int64
	s.DB.Model(&model.ChatMessage{}).Where("thread_id = ?", threadID).Count(&total)
	if err := s.DB.Where("thread_id = ?", threadID).Order("id DESC").
		Offset(int((current - 1) * size)).Limit(int(size)).Find(&msgs).Error; err != nil {
		return nil, 0, err
	}
	return msgs, total, nil
}

func (s *Service) ChatMarkThreadRead(userID, threadID int64) error {
	var thread model.ChatThread
	if err := s.DB.First(&thread, threadID).Error; err != nil {
		return err
	}
	var last model.ChatMessage
	s.DB.Where("thread_id = ?", threadID).Order("id DESC").First(&last)
	if last.ID == 0 {
		return nil
	}
	if thread.UserLow == userID {
		return s.DB.Model(&thread).Update("user_low_read_msg_id", last.ID).Error
	}
	return s.DB.Model(&thread).Update("user_high_read_msg_id", last.ID).Error
}

func (s *Service) privateUnread(thread model.ChatThread, userID int64) int64 {
	var readID *int64
	if thread.UserLow == userID {
		readID = thread.UserLowReadMsgID
	} else {
		readID = thread.UserHighReadMsgID
	}
	q := s.DB.Model(&model.ChatMessage{}).Where("thread_id = ? AND sender_id <> ?", thread.ID, userID)
	if readID != nil {
		q = q.Where("id > ?", *readID)
	}
	var count int64
	q.Count(&count)
	return count
}

func (s *Service) ChatUnreadTotal(userID int64) (map[string]any, error) {
	var private int64
	var threads []model.ChatThread
	s.DB.Where("user_low = ? OR user_high = ?", userID, userID).Find(&threads)
	for _, t := range threads {
		private += s.privateUnread(t, userID)
	}
	var group int64
	var members []model.ChatGroupMember
	s.DB.Where("user_id = ? AND member_status = 0", userID).Find(&members)
	for _, m := range members {
		group += s.groupUnread(m.GroupID, userID)
	}
	notice := s.noticeUnread(userID)
	total := private + group + notice["likeCollect"] + notice["follow"] + notice["comment"]
	return map[string]any{
		"total":       total,
		"likeCollect": notice["likeCollect"],
		"follow":      notice["follow"],
		"comment":     notice["comment"],
	}, nil
}

func (s *Service) groupUnread(groupID, userID int64) int64 {
	var member model.ChatGroupMember
	if err := s.DB.Where("group_id = ? AND user_id = ?", groupID, userID).First(&member).Error; err != nil {
		return 0
	}
	q := s.DB.Model(&model.GroupMessage{}).Where("group_id = ? AND sender_id <> ?", groupID, userID)
	if member.LastReadMsgID != nil {
		q = q.Where("id > ?", *member.LastReadMsgID)
	}
	var count int64
	q.Count(&count)
	return count
}
