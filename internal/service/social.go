package service

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"redbook-go/internal/model"
	"redbook-go/internal/pkg/idgen"
)

type ViewResult struct {
	Clicked bool  `json:"clicked"`
	Cnt     int64 `json:"cnt"`
}

type ClickResult struct {
	Clicked bool `json:"clicked"`
}

func (s *Service) LikeView(userID, noteID int64) (ViewResult, error) {
	return ViewResult{Clicked: s.isLiked(userID, noteID), Cnt: s.likeCount(noteID)}, nil
}

func (s *Service) Like(userID, noteID int64) (ClickResult, error) {
	if s.likeMQ != nil {
		return s.likeMQ.PublishClick(userID, noteID, 1)
	}
	if s.isLiked(userID, noteID) {
		return ClickResult{Clicked: true}, nil
	}
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		record := model.ClickRecords{ID: idgen.NextID(), ReceiverID: noteID, ClickerID: userID, ClickType: 1, Created: time.Now()}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
		return upsertLikeSummary(tx, noteID, 1)
	})
	if err != nil {
		return ClickResult{Clicked: false}, err
	}
	return ClickResult{Clicked: true}, nil
}

func (s *Service) Unlike(userID, noteID int64) (ClickResult, error) {
	if s.likeMQ != nil {
		return s.likeMQ.PublishClick(userID, noteID, -1)
	}
	if !s.isLiked(userID, noteID) {
		return ClickResult{Clicked: true}, nil
	}
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		record := model.ClickRecords{ID: idgen.NextID(), ReceiverID: noteID, ClickerID: userID, ClickType: -1, Created: time.Now()}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
		return upsertLikeSummary(tx, noteID, -1)
	})
	if err != nil {
		return ClickResult{Clicked: false}, err
	}
	return ClickResult{Clicked: true}, nil
}

func (s *Service) isLiked(userID, noteID int64) bool {
	var record model.ClickRecords
	err := s.DB.Where("receiver_id = ? AND clicker_id = ?", noteID, userID).
		Order("created DESC").First(&record).Error
	return err == nil && record.ClickType == 1
}

func (s *Service) likeCount(noteID int64) int64 {
	var summary model.ClickSummary
	if err := s.DB.Where("receiver_id = ?", noteID).First(&summary).Error; err == nil {
		return summary.LikeCount
	}
	return 0
}

func upsertLikeSummary(tx *gorm.DB, noteID int64, delta int64) error {
	_, err := applyClickSummary(tx, noteID, delta)
	return err
}

func applyClickSummary(tx *gorm.DB, noteID int64, delta int64) (int64, error) {
	var summary model.ClickSummary
	err := tx.Where("receiver_id = ?", noteID).First(&summary).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		now := time.Now()
		summary = model.ClickSummary{ID: idgen.NextID(), ReceiverID: noteID, LikeCount: 0, Created: now, Modified: now}
		if err := tx.Create(&summary).Error; err != nil {
			return 0, err
		}
	} else if err != nil {
		return 0, err
	}
	newCount := summary.LikeCount + delta
	if newCount < 0 {
		newCount = 0
	}
	err = tx.Model(&model.ClickSummary{}).Where("receiver_id = ?", noteID).
		Updates(map[string]any{"like_count": newCount, "modified": time.Now()}).Error
	return newCount, err
}

func (s *Service) LikeMy(userID int64, current, size int64) ([]map[string]any, int64, error) {
	sub := s.DB.Model(&model.ClickRecords{}).Select("receiver_id").
		Where("clicker_id = ? AND click_type = 1", userID).Group("receiver_id")
	return s.notesByIDs(sub, current, size, &userID)
}

func (s *Service) LikeByUser(userID int64, viewerID *int64, current, size int64) ([]map[string]any, int64, error) {
	sub := s.DB.Model(&model.ClickRecords{}).Select("receiver_id").
		Where("clicker_id = ? AND click_type = 1", userID).Group("receiver_id")
	return s.notesByIDs(sub, current, size, viewerID)
}

func (s *Service) notesByIDs(sub *gorm.DB, current, size int64, viewerID *int64) ([]map[string]any, int64, error) {
	var total int64
	s.DB.Model(&model.Note{}).Where("id IN (?)", sub).Count(&total)
	var notes []model.Note
	if err := s.DB.Where("id IN (?)", sub).Order("create_time DESC").
		Offset(int((current - 1) * size)).Limit(int(size)).Find(&notes).Error; err != nil {
		return nil, 0, err
	}
	list := make([]map[string]any, 0, len(notes))
	for _, n := range notes {
		list = append(list, s.buildNoteVO(n, viewerID))
	}
	return list, total, nil
}

func (s *Service) Collect(userID, noteID int64) error {
	var existing model.NoteCollect
	err := s.DB.Where("user_id = ? AND note_id = ?", userID, noteID).First(&existing).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	err = s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&model.NoteCollect{UserID: userID, NoteID: noteID}).Error; err != nil {
			return err
		}
		return tx.Model(&model.Note{}).Where("id = ?", noteID).UpdateColumn("collect_count", gorm.Expr("collect_count + 1")).Error
	})
	return err
}

func (s *Service) Uncollect(userID, noteID int64) error {
	res := s.DB.Where("user_id = ? AND note_id = ?", userID, noteID).Delete(&model.NoteCollect{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		s.DB.Model(&model.Note{}).Where("id = ?", noteID).UpdateColumn("collect_count", gorm.Expr("GREATEST(collect_count - 1, 0)"))
	}
	return nil
}

func (s *Service) CollectMy(userID int64, current, size int64) ([]map[string]any, int64, error) {
	sub := s.DB.Model(&model.NoteCollect{}).Select("note_id").Where("user_id = ?", userID)
	return s.notesByIDs(sub, current, size, &userID)
}

func (s *Service) CollectByUser(userID int64, viewerID *int64, current, size int64) ([]map[string]any, int64, error) {
	sub := s.DB.Model(&model.NoteCollect{}).Select("note_id").Where("user_id = ?", userID)
	return s.notesByIDs(sub, current, size, viewerID)
}

func (s *Service) Follow(userID, followeeID int64) error {
	if userID == followeeID {
		return errors.New("不能关注自己")
	}
	var existing model.UserFollow
	err := s.DB.Where("follower_id = ? AND followee_id = ?", userID, followeeID).First(&existing).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&model.UserFollow{FollowerID: userID, FolloweeID: followeeID}).Error; err != nil {
			return err
		}
		tx.Model(&model.UserInfo{}).Where("user_id = ?", userID).UpdateColumn("followee", gorm.Expr("followee + 1"))
		tx.Model(&model.UserInfo{}).Where("user_id = ?", followeeID).UpdateColumn("fans", gorm.Expr("fans + 1"))
		return nil
	})
}

func (s *Service) Unfollow(userID, followeeID int64) error {
	res := s.DB.Where("follower_id = ? AND followee_id = ?", userID, followeeID).Delete(&model.UserFollow{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected > 0 {
		s.DB.Model(&model.UserInfo{}).Where("user_id = ?", userID).UpdateColumn("followee", gorm.Expr("GREATEST(followee - 1, 0)"))
		s.DB.Model(&model.UserInfo{}).Where("user_id = ?", followeeID).UpdateColumn("fans", gorm.Expr("GREATEST(fans - 1, 0)"))
	}
	return nil
}

func (s *Service) Followers(userID int64, current, size int64) ([]model.User, int64, error) {
	sub := s.DB.Model(&model.UserFollow{}).Select("follower_id").Where("followee_id = ?", userID)
	return s.usersByIDs(sub, current, size)
}

func (s *Service) Following(userID int64, current, size int64) ([]model.User, int64, error) {
	sub := s.DB.Model(&model.UserFollow{}).Select("followee_id").Where("follower_id = ?", userID)
	return s.usersByIDs(sub, current, size)
}

func (s *Service) usersByIDs(sub *gorm.DB, current, size int64) ([]model.User, int64, error) {
	var total int64
	s.DB.Model(&model.User{}).Where("id IN (?)", sub).Count(&total)
	var users []model.User
	if err := s.DB.Where("id IN (?)", sub).Order("id DESC").Offset(int((current - 1) * size)).Limit(int(size)).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}
