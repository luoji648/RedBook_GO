package service

import (
	"time"

	"redbook-go/internal/model"
)

func (s *Service) noticeUnread(userID int64) map[string]int64 {
	result := map[string]int64{"likeCollect": 0, "follow": 0, "comment": 0}
	readTime := s.lastReadTime(userID, "like_collect")
	var likeCount int64
	s.DB.Model(&model.ClickRecords{}).
		Joins("JOIN tb_note ON tb_note.id = click_records.receiver_id").
		Where("tb_note.user_id = ? AND click_records.clicker_id <> ? AND click_records.click_type = 1 AND click_records.created > ?", userID, userID, readTime).
		Count(&likeCount)
	var collectCount int64
	s.DB.Model(&model.NoteCollect{}).
		Joins("JOIN tb_note ON tb_note.id = tb_note_collect.note_id").
		Where("tb_note.user_id = ? AND tb_note_collect.user_id <> ? AND tb_note_collect.create_time > ?", userID, userID, readTime).
		Count(&collectCount)
	result["likeCollect"] = likeCount + collectCount

	readFollow := s.lastReadTime(userID, "follow")
	var followCount int64
	s.DB.Model(&model.UserFollow{}).Where("followee_id = ? AND follower_id <> ? AND create_time > ?", userID, userID, readFollow).
		Count(&followCount)
	result["follow"] = followCount

	readComment := s.lastReadTime(userID, "comment")
	var commentCount int64
	s.DB.Model(&model.NoteComment{}).
		Joins("JOIN tb_note ON tb_note.id = tb_note_comment.note_id").
		Where("tb_note.user_id = ? AND tb_note_comment.user_id <> ? AND tb_note_comment.create_time > ?", userID, userID, readComment).
		Count(&commentCount)
	result["comment"] = commentCount
	return result
}

func (s *Service) lastReadTime(userID int64, category string) time.Time {
	var nr model.NoticeRead
	if err := s.DB.Where("user_id = ? AND category = ?", userID, category).First(&nr).Error; err == nil {
		return nr.LastReadTime
	}
	return time.Unix(0, 0)
}

func (s *Service) NoticeMarkRead(userID int64, category string) error {
	if category != "like_collect" && category != "follow" && category != "comment" {
		return nil
	}
	var nr model.NoticeRead
	err := s.DB.Where("user_id = ? AND category = ?", userID, category).First(&nr).Error
	if err == nil {
		return s.DB.Model(&nr).Update("last_read_time", time.Now()).Error
	}
	return s.DB.Create(&model.NoticeRead{UserID: userID, Category: category, LastReadTime: time.Now()}).Error
}

func (s *Service) NoticeInteractionPage(userID int64, current, size int64, category string) ([]map[string]any, int64, error) {
	// The interaction feed is a lightweight compatibility implementation. It
	// returns normalized rows from likes, collects, follows, and comments.
	type row struct {
		Category  string
		UserID    int64
		NoteID    *int64
		CreatedAt time.Time
		Text      string
	}
	rows := []row{}

	if category == "all" || category == "like_collect" {
		var likes []model.ClickRecords
		s.DB.Raw(`SELECT cr.* FROM click_records cr
			JOIN tb_note n ON n.id = cr.receiver_id
			WHERE n.user_id = ? AND cr.clicker_id <> ? AND cr.click_type = 1
			ORDER BY cr.created DESC LIMIT 200`, userID, userID).Scan(&likes)
		for _, l := range likes {
			rows = append(rows, row{Category: "like_collect", UserID: l.ClickerID, NoteID: &l.ReceiverID, CreatedAt: l.Created, Text: "赞了你的笔记"})
		}
		var collects []model.NoteCollect
		s.DB.Raw(`SELECT c.* FROM tb_note_collect c
			JOIN tb_note n ON n.id = c.note_id
			WHERE n.user_id = ? AND c.user_id <> ?
			ORDER BY c.create_time DESC LIMIT 200`, userID, userID).Scan(&collects)
		for _, c := range collects {
			rows = append(rows, row{Category: "like_collect", UserID: c.UserID, NoteID: &c.NoteID, CreatedAt: c.CreateTime, Text: "收藏了你的笔记"})
		}
	}
	if category == "all" || category == "follow" {
		var follows []model.UserFollow
		s.DB.Where("followee_id = ? AND follower_id <> ?", userID, userID).Order("create_time DESC").Limit(200).Find(&follows)
		for _, f := range follows {
			rows = append(rows, row{Category: "follow", UserID: f.FollowerID, CreatedAt: f.CreateTime, Text: "关注了你"})
		}
	}
	if category == "all" || category == "comment" {
		var comments []model.NoteComment
		s.DB.Raw(`SELECT c.* FROM tb_note_comment c
			JOIN tb_note n ON n.id = c.note_id
			WHERE n.user_id = ? AND c.user_id <> ?
			ORDER BY c.create_time DESC LIMIT 200`, userID, userID).Scan(&comments)
		for _, c := range comments {
			rows = append(rows, row{Category: "comment", UserID: c.UserID, NoteID: &c.NoteID, CreatedAt: c.CreateTime, Text: "评论了你的笔记"})
		}
	}

	total := int64(len(rows))
	start := int((current - 1) * size)
	if start > len(rows) {
		start = len(rows)
	}
	end := start + int(size)
	if end > len(rows) {
		end = len(rows)
	}
	list := make([]map[string]any, 0, end-start)
	for _, r := range rows[start:end] {
		var user model.User
		s.DB.First(&user, r.UserID)
		m := map[string]any{"category": r.Category, "user": user, "text": r.Text, "createdAt": r.CreatedAt}
		if r.NoteID != nil {
			m["noteId"] = *r.NoteID
		}
		list = append(list, m)
	}
	return list, total, nil
}
