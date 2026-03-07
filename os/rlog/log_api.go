package rlog

import "context"

func (l *Logger) Trace(ctx context.Context, msg string, args ...any) {
	l.withCtx(ctx).Trace(append([]any{msg}, args...)...)
}

func (l *Logger) Tracef(ctx context.Context, format string, v ...any) {
	l.withCtx(ctx).Tracef(format, v...)
}

func (l *Logger) Debug(ctx context.Context, msg string, args ...any) {
	l.withCtx(ctx).Debug(append([]any{msg}, args...)...)
}

func (l *Logger) Debugf(ctx context.Context, format string, v ...any) {
	l.withCtx(ctx).Debugf(format, v...)
}

func (l *Logger) Info(ctx context.Context, msg string, args ...any) {
	l.withCtx(ctx).Info(append([]any{msg}, args...)...)
}

func (l *Logger) Infof(ctx context.Context, format string, v ...any) {
	l.withCtx(ctx).Infof(format, v...)
}

func (l *Logger) Notice(ctx context.Context, msg string, args ...any) {
	l.withCtx(ctx).Notice(append([]any{msg}, args...)...)
}

func (l *Logger) Noticef(ctx context.Context, format string, v ...any) {
	l.withCtx(ctx).Noticef(format, v...)
}

func (l *Logger) Warn(ctx context.Context, msg string, args ...any) {
	l.withCtx(ctx).Warn(append([]any{msg}, args...)...)
}

func (l *Logger) Warnf(ctx context.Context, format string, v ...any) {
	l.withCtx(ctx).Warnf(format, v...)
}

func (l *Logger) Error(ctx context.Context, msg string, args ...any) {
	l.withCtx(ctx).Error(append([]any{msg}, args...)...)
}

func (l *Logger) Errorf(ctx context.Context, format string, v ...any) {
	l.withCtx(ctx).Errorf(format, v...)
}

func (l *Logger) Fatal(ctx context.Context, msg string, args ...any) {
	l.withCtx(ctx).Fatal(append([]any{msg}, args...)...)
}

func (l *Logger) Fatalf(ctx context.Context, format string, v ...any) {
	l.withCtx(ctx).Fatalf(format, v...)
}

func (l *Logger) Panic(ctx context.Context, msg string, args ...any) {
	l.withCtx(ctx).Panic(append([]any{msg}, args...)...)
}

func (l *Logger) Panicf(ctx context.Context, format string, v ...any) {
	l.withCtx(ctx).Panicf(format, v...)
}

// Print 兼容接口，不接受 ctx。
func (l *Logger) Print(v ...any) {
	l.logger.Info(v...)
}

// Printf 兼容接口，不接受 ctx。
func (l *Logger) Printf(format string, v ...any) {
	l.logger.Infof(format, v...)
}
