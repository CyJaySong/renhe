package rlog

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/gookit/slog"
	"github.com/uptrace/bun"
)

type bunQueryHook struct {
	l                  *Logger
	slowQueryThreshold time.Duration
}

// BunQueryHook 返回实现了 bun.QueryHook 接口的适配器，用于记录 SQL 查询日志。
// slowThreshold 为慢查询阈值，为 0 时不做慢查询判断。
func (l *Logger) BunQueryHook(slowThreshold ...time.Duration) *bunQueryHook {
	h := &bunQueryHook{l: l}
	if len(slowThreshold) > 0 {
		h.slowQueryThreshold = slowThreshold[0]
	}
	return h
}

func (h *bunQueryHook) BeforeQuery(ctx context.Context, _ *bun.QueryEvent) context.Context {
	return ctx
}

func (h *bunQueryHook) AfterQuery(ctx context.Context, event *bun.QueryEvent) {
	dur := time.Since(event.StartTime)

	if event.Err != nil && !errors.Is(event.Err, sql.ErrNoRows) {
		h.l.withCtx(ctx).WithData(slog.M{
			"op":       event.Operation(),
			"query":    event.Query,
			"err":      event.Err.Error(),
			"duration": dur.String(),
		}).Error("[BUN]")
		return
	}

	if h.slowQueryThreshold > 0 && dur >= h.slowQueryThreshold {
		h.l.withCtx(ctx).WithData(slog.M{
			"op":       event.Operation(),
			"query":    event.Query,
			"duration": dur.String(),
		}).Warn("[BUN-SLOW]")
		return
	}

	h.l.withCtx(ctx).WithData(slog.M{
		"op":       event.Operation(),
		"query":    event.Query,
		"duration": dur.String(),
	}).Debug("[BUN]")
}
