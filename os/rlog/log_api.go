package rlog

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

// stackSkip 调用栈跳过层数：runtime.Callers → captureStack → appendStack → log_api method
const stackSkip = 4

func (l *Logger) appendStack(msg any) string {
	var returnMsg string
	switch m := msg.(type) {
	case error:
		returnMsg = m.Error()
	case string:
		returnMsg = m
	default:
		returnMsg = fmt.Sprintf("%v", msg)
	}
	if !l.stack {
		return returnMsg
	}
	if s := captureStack(stackSkip); s != "" {
		return returnMsg + "\n" + s
	}
	return returnMsg
}

func (l *Logger) appendStackf(format string, v ...any) string {
	if !l.stack {
		return fmt.Sprintf(format, v...)
	}
	msg := fmt.Sprintf(format, v...)
	if s := captureStack(stackSkip); s != "" {
		return msg + "\n" + s
	}
	return msg
}

// msgString 将任意消息转为字符串。
func msgString(msg any) string {
	switch m := msg.(type) {
	case string:
		return m
	case error:
		return m.Error()
	default:
		return fmt.Sprintf("%v", m)
	}
}

// log 统一写日志：注入 ctx 字段，args 为 key-value 交替参数。
func (l *Logger) log(ctx context.Context, level slog.Level, msg string, args ...any) {
	if ctx == nil {
		ctx = context.Background()
	}
	// 未启用该级别则直接返回
	if !l.logger.Enabled(ctx, level) {
		return
	}
	// 合并 context 附加字段
	fields := l.ctxFields(ctx)
	if len(fields) == 0 {
		l.logger.Log(ctx, level, msg, args...)
		return
	}
	// With 附加字段后再打日志
	kv := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		kv = append(kv, k, v)
	}
	l.logger.With(kv...).Log(ctx, level, msg, args...)
}

func (l *Logger) Debug(ctx context.Context, msg any, args ...any) {
	l.log(ctx, slog.LevelDebug, msgString(msg), args...)
}

func (l *Logger) Debugf(ctx context.Context, format string, v ...any) {
	l.log(ctx, slog.LevelDebug, fmt.Sprintf(format, v...))
}

func (l *Logger) Info(ctx context.Context, msg any, args ...any) {
	l.log(ctx, slog.LevelInfo, msgString(msg), args...)
}

func (l *Logger) Infof(ctx context.Context, format string, v ...any) {
	l.log(ctx, slog.LevelInfo, fmt.Sprintf(format, v...))
}

func (l *Logger) Warn(ctx context.Context, msg any, args ...any) {
	l.log(ctx, slog.LevelWarn, msgString(msg), args...)
}

func (l *Logger) Warnf(ctx context.Context, format string, v ...any) {
	l.log(ctx, slog.LevelWarn, fmt.Sprintf(format, v...))
}

func (l *Logger) Error(ctx context.Context, msg any, args ...any) {
	l.log(ctx, slog.LevelError, l.appendStack(msg), args...)
}

func (l *Logger) Errorf(ctx context.Context, format string, v ...any) {
	l.log(ctx, slog.LevelError, l.appendStackf(format, v...))
}

// Fatal 以 Error 级别写日志后 os.Exit(1)。
func (l *Logger) Fatal(ctx context.Context, msg any, args ...any) {
	l.log(ctx, slog.LevelError, l.appendStack(msg), args...)
	os.Exit(1)
}

func (l *Logger) Fatalf(ctx context.Context, format string, v ...any) {
	l.log(ctx, slog.LevelError, l.appendStackf(format, v...))
	os.Exit(1)
}

// Panic 以 Error 级别写日志后 panic。
func (l *Logger) Panic(ctx context.Context, msg any, args ...any) {
	m := l.appendStack(msg)
	l.log(ctx, slog.LevelError, m, args...)
	panic(m)
}

func (l *Logger) Panicf(ctx context.Context, format string, v ...any) {
	msg := l.appendStackf(format, v...)
	l.log(ctx, slog.LevelError, msg)
	panic(msg)
}

// Print 兼容接口，不接受 ctx。
func (l *Logger) Print(v ...any) {
	l.log(context.Background(), slog.LevelInfo, fmt.Sprint(v...))
}

// Printf 兼容接口，不接受 ctx。
func (l *Logger) Printf(format string, v ...any) {
	l.log(context.Background(), slog.LevelInfo, fmt.Sprintf(format, v...))
}
