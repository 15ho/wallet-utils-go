package zlog

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func init() {
	if logger == nil {
		initLogger()
	}
}

func initLogger() {
	lvl := zap.DebugLevel
	if lvlStr := os.Getenv("GO_LOG"); lvlStr != "" {
		parsedLvl, err := zapcore.ParseLevel(os.Getenv("GO_LOG"))
		if err != nil {
			panic("parse log level error:" + err.Error())
		}
		lvl = parsedLvl
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(lvl)
	cfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	logger, _ = cfg.Build()
}

func Debug(msg string, fields ...zap.Field) {
	logger.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	logger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}
