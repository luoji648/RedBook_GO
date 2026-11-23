package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"time"

	"redbook-go/internal/model"
)

func (s *Service) GenerateShortLink(noteID int64) (map[string]any, error) {
	var note model.Note
	if err := s.DB.First(&note, noteID).Error; err != nil {
		return nil, errors.New("笔记不存在")
	}
	shortURI := randomURI(6)
	longURL := "http://localhost:5173/note/" + itoa(noteID)
	valid := time.Now().Add(30 * 24 * time.Hour)
	link := model.ShortLink{ShortURI: &shortURI, LongURL: &longURL, ValidDate: &valid, DelFlag: 0}
	if err := s.DB.Create(&link).Error; err != nil {
		// Collision or unique violation fallback: try once more.
		shortURI = randomURI(6)
		link.ShortURI = &shortURI
		if err := s.DB.Create(&link).Error; err != nil {
			return nil, err
		}
	}
	return map[string]any{"id": link.ID, "shortUri": shortURI, "longUrl": longURL, "validDate": valid}, nil
}

func (s *Service) ResolveShortLink(code string) (string, error) {
	var link model.ShortLink
	if err := s.DB.Where("short_uri = ? AND del_flag = 0", code).First(&link).Error; err != nil {
		return "", errors.New("短链接不存在")
	}
	if link.ValidDate != nil && time.Now().After(*link.ValidDate) {
		return "", errors.New("短链接已过期")
	}
	if link.LongURL == nil {
		return "", errors.New("短链接无效")
	}
	return *link.LongURL, nil
}

func (s *Service) DifyStream(ctx context.Context, userID, noteID int64, mode, question string) (*http.Response, error) {
	var note model.Note
	if err := s.DB.First(&note, noteID).Error; err != nil {
		return nil, errors.New("笔记不存在")
	}
	postContent := "标题:" + note.Title + "\n正文:" + note.Content
	comments := ""
	if mode == "SUMMARY_COMMENTS" || mode == "QA" {
		var rows []model.NoteComment
		s.DB.Where("note_id = ? AND status = 0", noteID).Where("parent_id IS NULL OR parent_id = 0").
			Order("create_time ASC").Limit(10).Find(&rows)
		for _, c := range rows {
			if c.Content != "" {
				comments += c.Content + "\n"
			}
		}
	}
	query := "请总结上述帖子内容"
	if mode == "SUMMARY_COMMENTS" {
		query = "请总结上述评论区内容"
	} else if mode == "QA" && question != "" {
		query = question
	}
	inputs := map[string]string{
		"mode":         mode,
		"post_content": postContent,
		"comments":     comments,
		"note_context": postContent + "\n评论区:\n" + comments,
	}
	return s.Dify.Stream(ctx, query, userID, inputs)
}

func randomURI(n int) string {
	const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	raw := make([]byte, n)
	_, _ = rand.Read(raw)
	for i := range raw {
		b[i] = alphabet[int(raw[i])%len(alphabet)]
	}
	return string(b)
}

func hexID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
