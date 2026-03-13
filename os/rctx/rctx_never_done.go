package rctx

import (
	"context"
	"time"
)

// neverDoneCtx wraps a context to never be done, while preserving its values.
// This is useful for passing trace context to async goroutines without cancellation.
type neverDoneCtx struct {
	context.Context
}

func (neverDoneCtx) Deadline() (deadline time.Time, ok bool) { return }
func (neverDoneCtx) Done() <-chan struct{}                    { return nil }
func (neverDoneCtx) Err() error                              { return nil }

// NeverDone wraps the given context so that it is never cancelled or timed out.
// Values (including trace span) from the original context are preserved.
func NeverDone(ctx context.Context) context.Context {
	return neverDoneCtx{ctx}
}
