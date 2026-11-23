package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func Init(development bool) error {
	cfg := zap.NewProductionConfig()
	if development {
		cfg = zap.NewDevelopmentConfig()
	}
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, err := cfg.Build()
	if err != nil {
		return err
	}
	Log = logger
	return nil
}

func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}
