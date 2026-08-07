package rlog

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

// ctxKey 用于在 context 中存储日志附加字段的键。
type ctxKey struct{}

// WithFields 向 ctx 中注入日志附加字段，后续通过该 ctx 打印日志时会自动携带这些字段。
// 会拷贝 fields，避免调用方后续修改 map 影响日志。
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
	// 拷贝入参，避免外部持有同一 map 引用
	copied := make(map[string]any, len(fields))
	for k, v := range fields {
		copied[k] = v
	}
	return context.WithValue(ctx, ctxKey{}, copied)
}

// ctxFields 从 ctx 中提取用户注入的字段和 OpenTelemetry Trace 信息。
func (l *Logger) ctxFields(ctx context.Context) map[string]any {
	if ctx == nil {
		return nil
	}
	fields, _ := ctx.Value(ctxKey{}).(map[string]any)
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return fields
	}
	sc := span.SpanContext()
	merged := make(map[string]any, len(fields)+2)
	for k, v := range fields {
		merged[k] = v
	}
	merged["traceId"] = sc.TraceID().String()
	merged["spanId"] = sc.SpanID().String()
	return merged
}
