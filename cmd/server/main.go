package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"go.uber.org/zap"

	"redbook-go/internal/auth"
	"redbook-go/internal/config"
	"redbook-go/internal/database"
	"redbook-go/internal/pkg/dify"
	"redbook-go/internal/pkg/logger"
	"redbook-go/internal/pkg/oss"
	"redbook-go/internal/router"
	"redbook-go/internal/service"
)

func main() {
	if err := logger.Init(false); err != nil {
		fmt.Fprintln(os.Stderr, "init logger:", err)
		os.Exit(1)
	}
	defer logger.Sync()

	cfg, err := config.Load()
	if err != nil {
		logger.Log.Fatal("load config", zap.Error(err))
	}

	store, err := database.Open(cfg)
	if err != nil {
		logger.Log.Fatal("open database", zap.Error(err))
	}
	if err := store.AutoMigrate(); err != nil {
		logger.Log.Fatal("auto migrate", zap.Error(err))
	}

	tokens := auth.NewTokenService(cfg.JWT.Secret, cfg.JWT.ExpireSeconds, store.Redis)
	ossClient := oss.New(cfg.OSS)
	difyClient := dify.New(cfg.Dify.URL, cfg.Dify.NoteAssistantKey)
	svc := service.New(store.DB, store.Redis, tokens, cfg, ossClient, difyClient)
	if cfg.RocketMQ.Enabled {
		likeMQ, err := service.NewLikeMQ(cfg, store.DB, store.Redis)
		if err != nil {
			logger.Log.Fatal("init like mq", zap.Error(err))
		}
		defer likeMQ.Stop()
		svc.AttachLikeMQ(likeMQ)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc.StartSchedulers(ctx, 0)

	engine := router.New(svc)
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Log.Info("redbook-go backend listening", zap.String("addr", addr))
	if err := http.ListenAndServe(addr, engine); err != nil {
		logger.Log.Fatal("server stopped", zap.Error(err))
	}
}
