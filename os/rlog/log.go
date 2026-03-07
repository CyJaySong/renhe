package rlog

import (
	"fmt"
	"strings"
	"sync"

	"github.com/cyjaysong/renhe/os/rcfg"
	"github.com/gookit/slog"
	"github.com/gookit/slog/handler"
)

var (
	instance *Logger
	once     sync.Once
)

type Logger struct {
	logger *slog.Logger
}

func Log() *Logger {
	once.Do(func() {
		level := loadLevel()
		inner := slog.New()
		inner.ReportCaller, inner.CallerSkip = true, 4
		inner.AddHandler(handler.ConsoleWithMaxLevel(level))
		instance = &Logger{logger: inner}
	})
	return instance
}

func loadLevel() slog.Level {
	v := rcfg.Cfg()
	key := "logger.level"
	switch strings.ToLower(v.GetString(key)) {
	case "trace":
		return slog.TraceLevel
	case "debug":
		return slog.DebugLevel
	case "info":
		return slog.InfoLevel
	case "notice":
		return slog.NoticeLevel
	case "warn", "warning":
		return slog.WarnLevel
	case "error":
		return slog.ErrorLevel
	case "fatal":
		return slog.FatalLevel
	default:
		return slog.InfoLevel
	}
}

// Inner 返回底层 *slog.Logger，便于高级用法。
func (l *Logger) Inner() *slog.Logger {
	return l.logger
}

// AddHandler 为当前 Logger 追加 Handler。
func (l *Logger) AddHandler(h slog.Handler) {
	l.logger.AddHandler(h)
}

// M String 格式化辅助
func M(key string, val any) string {
	return fmt.Sprintf("%s=%v", key, val)
}
