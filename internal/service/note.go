package service

import (
	"errors"
	"time"

	"gorm.io/gorm"

	"redbook-go/internal/model"
)

type NoteSaveRequest struct {
	ID         *int64  `json:"id"`
	Title      string  `json:"title"`
	Content    string  `json:"content"`
	Type       int32   `json:"type"`
	Visibility int32   `json:"visibility"`
	MediaURLs  []string `json:"mediaUrls"`
	ProductIDs []int64 `json:"productIds"`
	Publish    bool    `json:"publish"`
}

type CommentRequest struct {
	ParentID      *int64 `json:"parentId"`
	ReplyToUserID *int64 `json:"replyToUserId"`
	Content       string `json:"content"`
}

func (s *Service) SaveNote(userID int64, req NoteSaveRequest) (map[string]any, error) {
	status := int32(0)
	if req.Publish {
		status = 1
	}

	var note model.Note
	if req.ID != nil {
		if err := s.DB.Where("id = ? AND user_id = ?", *req.ID, userID).First(&note).Error; err != nil {
			return nil, errors.New("笔记不存在")
		}
		note.Title = req.Title
		note.Content = req.Content
		note.Type = req.Type
		note.Visibility = req.Visibility
		if status == 1 || note.Status != 1 {
			note.Status = status
		}
		note.UpdateTime = time.Now()
		if err := s.DB.Save(&note).Error; err != nil {
			return nil, err
		}
		s.DB.Where("note_id = ?", note.ID).Delete(&model.NoteMedia{})
		s.DB.Where("note_id = ?", note.ID).Delete(&model.NoteProduct{})
	} else {
		note = model.Note{
			UserID:     userID,
			Title:      req.Title,
			Content:    req.Content,
			Type:       req.Type,
			Status:     status,
			Visibility: req.Visibility,
		}
		if err := s.DB.Create(&note).Error; err != nil {
			return nil, err
		}
	}

	for i, url := range req.MediaURLs {
		s.DB.Create(&model.NoteMedia{NoteID: note.ID, URL: url, SortOrder: int32(i), MediaType: req.Type})
	}
	for i, productID := range req.ProductIDs {
		s.DB.Create(&model.NoteProduct{NoteID: note.ID, ProductID: productID, SortOrder: int32(i)})
	}
	return s.NoteDetail(note.ID, &userID)
}

func (s *Service) PublishNote(userID, id int64) (map[string]any, error) {
	res := s.DB.Model(&model.Note{}).Where("id = ? AND user_id = ?", id, userID).Update("status", 1)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, errors.New("笔记不存在")
	}
	return s.NoteDetail(id, &userID)
}

func (s *Service) DeleteNote(userID, id int64) error {
	res := s.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&model.Note{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("笔记不存在")
	}
	return nil
}

func (s *Service) NoteDetail(id int64, viewerID *int64) (map[string]any, error) {
	var note model.Note
	if err := s.DB.First(&note, id).Error; err != nil {
		return nil, errors.New("笔记不存在")
	}
	return s.buildNoteVO(note, viewerID), nil
}

func (s *Service) MyNotes(userID int64, current, size int64) ([]map[string]any, int64, error) {
	var notes []model.Note
	var total int64
	q := s.DB.Where("user_id = ?", userID)
	q.Model(&model.Note{}).Count(&total)
	if err := q.Order("create_time DESC").Offset(int((current - 1) * size)).Limit(int(size)).Find(&notes).Error; err != nil {
		return nil, 0, err
	}
	list := make([]map[string]any, 0, len(notes))
	for _, n := range notes {
		list = append(list, s.buildNoteVO(n, &userID))
	}
	return list, total, nil
}

func (s *Service) Feed(userID int64, maxTime *int64, offset *int, size int64) (map[string]any, error) {
	sub := s.DB.Model(&model.UserFollow{}).Select("followee_id").Where("follower_id = ?", userID)
	q := s.DB.Where("status = ?", 1).Where("user_id IN (?)", sub).Order("create_time DESC")
	if maxTime != nil {
		q = q.Where("create_time < ?", time.UnixMilli(*maxTime))
	}
	var notes []model.Note
	if err := q.Limit(int(size) + 1).Find(&notes).Error; err != nil {
		return nil, err
	}
	list := make([]map[string]any, 0, len(notes))
	var minTime int64
	for _, n := range notes {
		list = append(list, s.buildNoteVO(n, &userID))
		minTime = n.CreateTime.UnixMilli()
	}
	return map[string]any{"list": list, "minTime": minTime, "offset": offset}, nil
}

func (s *Service) Recommend(current, size int64) ([]map[string]any, int64, error) {
	var notes []model.Note
	var total int64
	s.DB.Model(&model.Note{}).Where("status = ? AND visibility = 0", 1).Count(&total)
	if err := s.DB.Where("status = ? AND visibility = 0", 1).Order("create_time DESC").
		Offset(int((current - 1) * size)).Limit(int(size)).Find(&notes).Error; err != nil {
		return nil, 0, err
	}
	list := make([]map[string]any, 0, len(notes))
	for _, n := range notes {
		list = append(list, s.buildNoteVO(n, nil))
	}
	return list, total, nil
}

func (s *Service) Related(id int64) ([]map[string]any, error) {
	var notes []model.Note
	if err := s.DB.Where("status = ? AND visibility = 0 AND id <> ?", 1, id).
		Order("create_time DESC").Limit(10).Find(&notes).Error; err != nil {
		return nil, err
	}
	list := make([]map[string]any, 0, len(notes))
	for _, n := range notes {
		list = append(list, s.buildNoteVO(n, nil))
	}
	return list, nil
}

func (s *Service) UserNotes(target, viewerID *int64, current, size int64) ([]map[string]any, int64, error) {
	q := s.DB.Model(&model.Note{}).Where("user_id = ?", *target)
	if viewerID == nil || *viewerID != *target {
		q = q.Where("visibility = 0")
	} else {
		q = q.Where("user_id = ?", *target)
	}
	var total int64
	q.Count(&total)
	var notes []model.Note
	if err := q.Where("status = ?", 1).Order("create_time DESC").Offset(int((current - 1) * size)).Limit(int(size)).Find(&notes).Error; err != nil {
		return nil, 0, err
	}
	list := make([]map[string]any, 0, len(notes))
	for _, n := range notes {
		list = append(list, s.buildNoteVO(n, viewerID))
	}
	return list, total, nil
}

func (s *Service) buildNoteVO(note model.Note, viewerID *int64) map[string]any {
	var media []model.NoteMedia
	s.DB.Where("note_id = ?", note.ID).Order("sort_order ASC").Find(&media)

	var products []model.NoteProduct
	s.DB.Where("note_id = ?", note.ID).Order("sort_order ASC").Find(&products)

	var author model.User
	s.DB.First(&author, note.UserID)

	likeCount := s.likeCount(note.ID)
	vo := map[string]any{
		"note":         note,
		"author":       author,
		"media":        media,
		"noteProducts": products,
		"likeCount":    likeCount,
	}
	if viewerID != nil {
		vo["liked"] = s.isLiked(*viewerID, note.ID)
		var collectCount int64
		s.DB.Model(&model.NoteCollect{}).Where("user_id = ? AND note_id = ?", *viewerID, note.ID).Count(&collectCount)
		vo["collected"] = collectCount > 0
	}
	return vo
}

func (s *Service) AddComment(noteID, userID int64, req CommentRequest) (any, error) {
	var note model.Note
	if err := s.DB.First(&note, noteID).Error; err != nil {
		return nil, errors.New("笔记不存在")
	}
	if req.Content == "" {
		return nil, errors.New("评论内容不能为空")
	}
	comment := model.NoteComment{
		NoteID:        noteID,
		UserID:        userID,
		ParentID:      req.ParentID,
		ReplyToUserID: req.ReplyToUserID,
		Content:       req.Content,
	}
	if err := s.DB.Create(&comment).Error; err != nil {
		return nil, err
	}
	s.DB.Model(&model.Note{}).Where("id = ?", noteID).UpdateColumn("comment_count", gorm.Expr("comment_count + 1"))
	return comment, nil
}

func (s *Service) CommentTree(noteID int64) (any, error) {
	var comments []model.NoteComment
	if err := s.DB.Where("note_id = ? AND status = 0", noteID).Order("create_time ASC").Find(&comments).Error; err != nil {
		return nil, err
	}
	type Node struct {
		Comment model.NoteComment `json:"comment"`
		User    model.User        `json:"user"`
		Replies []Node            `json:"replies"`
	}
	byID := map[int64]*Node{}
	roots := []Node{}
	for _, c := range comments {
		var user model.User
		s.DB.First(&user, c.UserID)
		node := Node{Comment: c, User: user, Replies: []Node{}}
		byID[c.ID] = &node
	}
	for _, c := range comments {
		node := byID[c.ID]
		if c.ParentID != nil && *c.ParentID != 0 {
			if parent, ok := byID[*c.ParentID]; ok {
				parent.Replies = append(parent.Replies, *node)
				continue
			}
		}
		roots = append(roots, *node)
	}
	return roots, nil
}

func (s *Service) DeleteComment(userID, commentID int64) error {
	res := s.DB.Where("id = ? AND user_id = ?", commentID, userID).Delete(&model.NoteComment{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("评论不存在")
	}
	return nil
}
