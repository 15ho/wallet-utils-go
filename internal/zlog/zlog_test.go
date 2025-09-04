package zlog

import (
	"os"
	"testing"

	"go.uber.org/zap"
)

func TestZlog(t *testing.T) {
	Info("test info", zap.String("hello", "world"))
	Warn("test warn", zap.String("msg", "warn warn warn"), zap.Int("code", 404))
	Error("test error", zap.String("msg", "error error error"), zap.Any("code", 500))
	Debug("no debug log")
	os.Setenv("GO_LOG", "debug")
	initLogger()
	Debug("test debug")
}
