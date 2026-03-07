package rdb

import (
	"context"

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

// Reader 返回读操作目标：事务中返回事务，否则返回从库。
func (d *DB) Reader(ctx context.Context) bun.IDB {
	if tx, ok := TxFromContext(ctx); ok {
		return tx
	}
	return d.Slave()
}

// Writer 返回写操作目标：事务中返回事务，否则返回主库。
func (d *DB) Writer(ctx context.Context) bun.IDB {
	if tx, ok := TxFromContext(ctx); ok {
		return tx
	}
	return d.master
}
