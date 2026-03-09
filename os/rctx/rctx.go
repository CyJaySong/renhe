package rctx

import (
	"context"
	"go.opentelemetry.io/otel/trace"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type (
	Ctx    = context.Context // Ctx is a short name alias for context.Context.
	StrKey string            // StrKey is a type for warps basic type string as context key.
)

var (
	// initCtx is the context initialized from process environment.
	initCtx context.Context
)

func init() {
	// All environment key-value pairs.
	m := make(map[string]string)
	i := 0
	for _, s := range os.Environ() {
		i = strings.IndexByte(s, '=')
		if i == -1 {
			continue
		}
		m[s[0:i]] = s[i+1:]
	}
	// OpenTelemetry from environments.
	initCtx = otel.GetTextMapPropagator().Extract(
		context.Background(),
		propagation.MapCarrier(m),
	)
}

// New creates and returns a context which contains context id.
func New() context.Context {
	return WithSpan(context.Background(), "rctx.New")
}

// WithCtx creates and returns a context containing context id upon given parent context `ctx`.
//
// Deprecated: use WithSpan instead.
func WithCtx(ctx context.Context) context.Context {
	if CtxId(ctx) != "" {
		return ctx
	}
	var span trace.Span
	ctx, span = otel.Tracer("").Start(ctx, "rctx.WithCtx")
	defer span.End()
	return ctx
}

// WithSpan creates and returns a context containing span upon given parent context `ctx`.
func WithSpan(ctx context.Context, spanName string) context.Context {
	if CtxId(ctx) != "" {
		return ctx
	}
	if spanName == "" {
		spanName = "rctx.WithSpan"
	}
	ctx, span := otel.Tracer(spanName).Start(ctx, spanName)
	defer span.End()
	return ctx
}

// CtxId retrieves and returns the context id from context.
func CtxId(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if traceID := trace.SpanContextFromContext(ctx).TraceID(); traceID.IsValid() {
		return traceID.String()
	}
	return ""
}

// SetInitCtx sets custom initialization context.
// Note that this function cannot be called in multiple goroutines.
func SetInitCtx(ctx context.Context) {
	initCtx = ctx
}

// GetInitCtx returns the initialization context.
// Initialization context is used in `main` or `init` functions.
func GetInitCtx() context.Context {
	return initCtx
}
