package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"redbook-go/internal/config"
	"redbook-go/internal/model"
)

type Store struct {
	DB    *gorm.DB
	Redis *redis.Client
}

func Open(cfg *config.Config) (*Store, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		cfg.MySQL.User, cfg.MySQL.Password, cfg.MySQL.Host, cfg.MySQL.Port, cfg.MySQL.Database)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		// Redis is required by the original auth flow, but development can continue
		// with in-memory fallbacks handled by callers.
		return &Store{DB: db, Redis: nil}, nil
	}

	return &Store{DB: db, Redis: rdb}, nil
}

func (s *Store) AutoMigrate() error {
	return s.DB.AutoMigrate(
		&model.User{},
		&model.UserInfo{},
		&model.Note{},
		&model.NoteMedia{},
		&model.NoteProduct{},
		&model.NoteComment{},
		&model.NoteCollect{},
		&model.UserFollow{},
		&model.Product{},
		&model.Cart{},
		&model.ShopOrder{},
		&model.OrderItem{},
		&model.UserCoupon{},
		&model.ProductFootprint{},
		&model.UserWallet{},
		&model.WalletRecharge{},
		&model.ChatThread{},
		&model.ChatMessage{},
		&model.ChatGroup{},
		&model.ChatGroupMember{},
		&model.ChatGroupJoinRequest{},
		&model.GroupMessage{},
		&model.NoticeRead{},
		&model.ShortLink{},
		&model.ClickRecords{},
		&model.ClickSummary{},
	)
}
