// Package rlog 提供全局日志服务，基于 gookit/slog 实现。
// 支持上下文字段注入、OpenTelemetry Trace 自动关联，并提供 echo.Logger 和 bun.QueryHook 适配器。
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

// Logger 封装 gookit/slog.Logger，提供带 context 的日志方法。
type Logger struct {
	logger *slog.Logger
}

// Log 返回全局 Logger 单例。首次调用时根据配置初始化日志级别和 Handler。
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

// loadLevel 从全局配置 logger.level 读取日志级别，默认 InfoLevel。
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
