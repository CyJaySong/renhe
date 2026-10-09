package rdb

import (
	"context"
	"database/sql"

	"github.com/uptrace/bun"
)

// txCtxKey 用于在 context 中存储当前事务的键。
type txCtxKey struct{}

// WithTx 将事务注入 context，后续通过该 ctx 的读写操作自动路由到事务。
func WithTx(ctx context.Context, tx bun.Tx) context.Context {
	return context.WithValue(ctx, txCtxKey{}, tx)
}

// TxFromContext 从 context 中提取事务（若存在）。
func TxFromContext(ctx context.Context) (bun.Tx, bool) {
	tx, ok := ctx.Value(txCtxKey{}).(bun.Tx)
	return tx, ok
}

// --- Transaction operations → always master ---

// BeginTx 开启事务并将其注入返回的 context，后续操作自动路由到事务。
func (d *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (tx bun.Tx, ctxWithTx context.Context, err error) {
	if tx, err = d.Writer(ctx).BeginTx(ctx, opts); err == nil {
		return tx, WithTx(ctx, tx), nil
	}
	return tx, ctx, err
}

// RunInTx 在事务中执行 fn，fn 收到的 ctx 已注入事务。fn 返回 error 时自动回滚，否则提交。
func (d *DB) RunInTx(ctx context.Context, opts *sql.TxOptions, fn func(ctx context.Context, tx bun.Tx) error) error {
	return d.Writer(ctx).RunInTx(ctx, opts, func(ctx context.Context, tx bun.Tx) error {
		return fn(WithTx(ctx, tx), tx)
	})
}
