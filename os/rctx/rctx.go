package rctx

import (
	"context"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
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

// New creates and returns a context which contains a trace id.
// The internal span is ended immediately; use WithSpan for real span tracking.
func New() context.Context {
	ctx, span := otel.Tracer("rctx").Start(context.Background(), "rctx.New")
	span.End()
	return ctx
}

// WithSpan creates a new span on the given ctx and returns both.
// Caller is responsible for calling span.End() when the operation completes.
//
//	ctx, span := rctx.WithSpan(ctx, "myOperation")
//	defer span.End()
func WithSpan(ctx context.Context, spanName string) (context.Context, trace.Span) {
	if spanName == "" {
		spanName = "rctx.WithSpan"
	}
	return otel.Tracer("rctx").Start(ctx, spanName)
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

// ReExtractInitCtx re-extracts OpenTelemetry context from environment variables.
// Should be called after otel.SetTextMapPropagator has been configured (e.g. after rotrace.Init).
func ReExtractInitCtx() {
	m := make(map[string]string)
	for _, s := range os.Environ() {
		if i := strings.IndexByte(s, '='); i != -1 {
			m[s[0:i]] = s[i+1:]
		}
	}
	initCtx = otel.GetTextMapPropagator().Extract(context.Background(), propagation.MapCarrier(m))
}

// GetInitCtx returns the initialization context.
// Initialization context is used in `main` or `init` functions.
func GetInitCtx() context.Context {
	return initCtx
}
