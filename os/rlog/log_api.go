package rlog

import (
	"context"
	"fmt"
	"os"
)

// stackSkip 调用栈跳过层数：runtime.Callers → captureStack → appendStack → log_api method
const stackSkip = 4

func (l *Logger) appendStack(msg any) string {
	var returnMsg string
	switch msg.(type) {
	case error:
		returnMsg = msg.(error).Error()
	case string:
		returnMsg = msg.(string)
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

func (l *Logger) Trace(ctx context.Context, msg any, args ...any) {
	l.withCtx(ctx).Trace(append([]any{msg}, args...)...)
}

func (l *Logger) Tracef(ctx context.Context, format string, v ...any) {
	l.withCtx(ctx).Tracef(format, v...)
}

func (l *Logger) Debug(ctx context.Context, msg any, args ...any) {
	l.withCtx(ctx).Debug(append([]any{msg}, args...)...)
}

func (l *Logger) Debugf(ctx context.Context, format string, v ...any) {
	l.withCtx(ctx).Debugf(format, v...)
}

func (l *Logger) Info(ctx context.Context, msg any, args ...any) {
	l.withCtx(ctx).Info(append([]any{msg}, args...)...)
}

func (l *Logger) Infof(ctx context.Context, format string, v ...any) {
	l.withCtx(ctx).Infof(format, v...)
}

func (l *Logger) Notice(ctx context.Context, msg any, args ...any) {
	l.withCtx(ctx).Notice(append([]any{msg}, args...)...)
}

func (l *Logger) Noticef(ctx context.Context, format string, v ...any) {
	l.withCtx(ctx).Noticef(format, v...)
}

func (l *Logger) Warn(ctx context.Context, msg any, args ...any) {
	l.withCtx(ctx).Warn(append([]any{msg}, args...)...)
}

func (l *Logger) Warnf(ctx context.Context, format string, v ...any) {
	l.withCtx(ctx).Warnf(format, v...)
}

func (l *Logger) Error(ctx context.Context, msg any, args ...any) {
	l.withCtx(ctx).Error(append([]any{l.appendStack(msg)}, args...)...)
}

func (l *Logger) Errorf(ctx context.Context, format string, v ...any) {
	l.withCtx(ctx).Error(l.appendStackf(format, v...))
}

func (l *Logger) Fatal(ctx context.Context, msg any, args ...any) {
	l.withCtx(ctx).Fatal(append([]any{l.appendStack(msg)}, args...)...)
	os.Exit(1)
}

func (l *Logger) Fatalf(ctx context.Context, format string, v ...any) {
	l.withCtx(ctx).Fatal(l.appendStackf(format, v...))
	os.Exit(1)
}

func (l *Logger) Panic(ctx context.Context, msg any, args ...any) {
	l.withCtx(ctx).Panic(append([]any{l.appendStack(msg)}, args...)...)
	panic(msg)
}

func (l *Logger) Panicf(ctx context.Context, format string, v ...any) {
	msg := l.appendStackf(format, v...)
	l.withCtx(ctx).Panic(msg)
	panic(msg)
}

// Print 兼容接口，不接受 ctx。
func (l *Logger) Print(v ...any) {
	l.logger.Info(v...)
}

// Printf 兼容接口，不接受 ctx。
func (l *Logger) Printf(format string, v ...any) {
	l.logger.Infof(format, v...)
}
