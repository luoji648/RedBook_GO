package service

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"redbook-go/internal/auth"
	"redbook-go/internal/config"
	"redbook-go/internal/pkg/dify"
	"redbook-go/internal/pkg/oss"
)

type Service struct {
	DB     *gorm.DB
	Redis  *redis.Client
	Tokens *auth.TokenService
	Cfg    *config.Config
	OSS    *oss.Client
	Dify   *dify.Client
	likeMQ *LikeMQ
}

func New(db *gorm.DB, redis *redis.Client, tokens *auth.TokenService, cfg *config.Config, ossClient *oss.Client, difyClient *dify.Client) *Service {
	return &Service{DB: db, Redis: redis, Tokens: tokens, Cfg: cfg, OSS: ossClient, Dify: difyClient}
}

func (s *Service) AttachLikeMQ(lm *LikeMQ) {
	s.likeMQ = lm
}
