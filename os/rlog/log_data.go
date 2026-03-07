package rlog

import (
	"context"

	"github.com/gookit/slog"
	"go.opentelemetry.io/otel/trace"
)

type ctxKey struct{}

// WithFields 向 ctx 中注入日志附加字段，后续通过该 ctx 打印日志时会自动携带这些字段。
func WithFields(ctx context.Context, fields map[string]any) context.Context {
	if existing, ok := ctx.Value(ctxKey{}).(map[string]any); ok {
		merged := make(map[string]any, len(existing)+len(fields))
		for k, v := range existing {
			merged[k] = v
		}
		for k, v := range fields {
			merged[k] = v
		}
		return context.WithValue(ctx, ctxKey{}, merged)
	}
	return context.WithValue(ctx, ctxKey{}, fields)
}

func (l *Logger) ctxFields(ctx context.Context) slog.M {
	fields, _ := ctx.Value(ctxKey{}).(map[string]any)
	if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
		sc := span.SpanContext()
		merged := make(slog.M, len(fields)+2)
		for k, v := range fields {
			merged[k] = v
		}
		merged["traceId"] = sc.TraceID().String()
		merged["spanId"] = sc.SpanID().String()
		return merged
	}
	return fields
}

func (l *Logger) withCtx(ctx context.Context) *slog.Record {
	r := l.logger.WithFields(l.ctxFields(ctx)).SetContext(ctx)
	return r
}
