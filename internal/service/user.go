package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"gorm.io/gorm"

	"redbook-go/internal/auth"
	"redbook-go/internal/model"
)

var memorySMS sync.Map
var memorySign sync.Map

func (s *Service) SendCode(phone string) (string, error) {
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	key := "redbook:sms:code:" + phone
	if s.Redis != nil {
		if err := s.Redis.Set(context.Background(), key, code, 5*time.Minute).Err(); err != nil {
			return "", err
		}
	} else {
		memorySMS.Store(key, code)
	}
	return code, nil
}

func (s *Service) LoginByCode(phone, code string) (string, error) {
	key := "redbook:sms:code:" + phone
	var expected string
	if s.Redis != nil {
		expected, _ = s.Redis.Get(context.Background(), key).Result()
	} else if v, ok := memorySMS.Load(key); ok {
		expected, _ = v.(string)
	}
	if expected == "" || expected != code {
		return "", errors.New("验证码错误")
	}
	user, err := s.getOrCreateUserByPhone(phone)
	if err != nil {
		return "", err
	}
	return s.Tokens.CreateToken(user.ID, user.TokenVersion)
}

func (s *Service) LoginByPassword(phone, password string) (string, error) {
	var user model.User
	if err := s.DB.Where("phone = ?", phone).First(&user).Error; err != nil {
		return "", errors.New("手机号或密码错误")
	}
	if !auth.MatchesPassword(user.Password, password) {
		return "", errors.New("手机号或密码错误")
	}
	return s.Tokens.CreateToken(user.ID, user.TokenVersion)
}

func (s *Service) Logout(authorization string) error {
	claims, err := s.Tokens.ParseToken(authorization)
	if err != nil || claims == nil || claims.ID == "" {
		return nil
	}
	s.Tokens.Blacklist(context.Background(), claims.ID, time.Duration(s.Cfg.JWT.ExpireSeconds)*time.Second)
	return nil
}

func (s *Service) Sign(userID int64) (map[string]any, error) {
	now := time.Now()
	key := fmt.Sprintf("redbook:sign:%d:%s", userID, now.Format("200601"))
	day := strconv.Itoa(now.Day())

	if s.Redis != nil {
		added, err := s.Redis.SAdd(context.Background(), key, day).Result()
		if err != nil {
			return nil, err
		}
		s.Redis.Expire(context.Background(), key, 40*24*time.Hour)
		return map[string]any{"firstSignToday": added > 0}, nil
	}

	m, _ := memorySign.LoadOrStore(key, &sync.Map{})
	set := m.(*sync.Map)
	_, loaded := set.LoadOrStore(day, struct{}{})
	return map[string]any{"firstSignToday": !loaded}, nil
}

func (s *Service) SignCount(userID int64) (int64, error) {
	now := time.Now()
	key := fmt.Sprintf("redbook:sign:%d:%s", userID, now.Format("200601"))
	if s.Redis != nil {
		return s.Redis.SCard(context.Background(), key).Result()
	}
	if m, ok := memorySign.Load(key); ok {
		var count int64
		m.(*sync.Map).Range(func(_, _ any) bool { count++; return true })
		return count, nil
	}
	return 0, nil
}

func (s *Service) Me(userID int64) (map[string]any, error) {
	var user model.User
	if err := s.DB.First(&user, userID).Error; err != nil {
		return nil, err
	}
	info, _ := s.getInfo(userID)
	return map[string]any{"user": user, "userInfo": info, "isSelf": true}, nil
}

func (s *Service) PublicProfile(userID int64, viewerID *int64) (map[string]any, error) {
	var user model.User
	if err := s.DB.First(&user, userID).Error; err != nil {
		return nil, errors.New("用户不存在")
	}
	info, _ := s.getInfo(userID)
	result := map[string]any{"user": user, "userInfo": info, "isSelf": false, "followedByMe": false}
	if viewerID != nil {
		result["isSelf"] = *viewerID == userID
		var count int64
		s.DB.Model(&model.UserFollow{}).Where("follower_id = ? AND followee_id = ?", *viewerID, userID).Count(&count)
		result["followedByMe"] = count > 0
	}
	return result, nil
}

func (s *Service) UpdateProfile(userID int64, nickName, icon string) (map[string]any, error) {
	updates := map[string]any{}
	if nickName != "" {
		updates["nick_name"] = nickName
	}
	if icon != "" {
		updates["icon"] = icon
	}
	if len(updates) > 0 {
		if err := s.DB.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	return s.Me(userID)
}

func (s *Service) MyInfo(userID int64) (*model.UserInfo, error) {
	return s.getInfo(userID)
}

func (s *Service) UpdateInfo(userID int64, updates map[string]any) (*model.UserInfo, error) {
	var existing model.UserInfo
	err := s.DB.Where("user_id = ?", userID).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		existing = model.UserInfo{UserID: userID}
		if err := s.DB.Create(&existing).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	if len(updates) > 0 {
		if err := s.DB.Model(&model.UserInfo{}).Where("user_id = ?", userID).Updates(updates).Error; err != nil {
			return nil, err
		}
	}
	return s.getInfo(userID)
}

func (s *Service) ChangePassword(userID int64, oldPassword, newPassword string) error {
	var user model.User
	if err := s.DB.First(&user, userID).Error; err != nil {
		return err
	}
	if !auth.MatchesPassword(user.Password, oldPassword) {
		return errors.New("原密码错误")
	}
	encoded := auth.EncodePassword(newPassword)
	return s.DB.Model(&model.User{}).Where("id = ?", userID).Updates(map[string]any{
		"password":      encoded,
		"token_version": gorm.Expr("token_version + 1"),
	}).Error
}

func (s *Service) getOrCreateUserByPhone(phone string) (*model.User, error) {
	var user model.User
	err := s.DB.Where("phone = ?", phone).First(&user).Error
	if err == nil {
		return &user, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	user = model.User{
		Phone:    phone,
		NickName: "用户" + phone[len(phone)-4:],
	}
	if err := s.DB.Create(&user).Error; err != nil {
		return nil, err
	}
	info := model.UserInfo{UserID: user.ID}
	_ = s.DB.Create(&info).Error
	wallet := model.UserWallet{UserID: user.ID}
	_ = s.DB.Create(&wallet).Error
	return &user, nil
}

func (s *Service) getInfo(userID int64) (*model.UserInfo, error) {
	var info model.UserInfo
	if err := s.DB.Where("user_id = ?", userID).First(&info).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			info = model.UserInfo{UserID: userID}
			_ = s.DB.Create(&info).Error
			return &info, nil
		}
		return nil, err
	}
	return &info, nil
}

