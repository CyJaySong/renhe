package rdb

import (
	"context"

	"github.com/uptrace/bun"
)

type txCtxKey struct{}

func WithTx(ctx context.Context, tx bun.Tx) context.Context {
	return context.WithValue(ctx, txCtxKey{}, tx)
}

func TxFromContext(ctx context.Context) (bun.Tx, bool) {
	tx, ok := ctx.Value(txCtxKey{}).(bun.Tx)
	return tx, ok
}

func (d *DB) Reader(ctx context.Context) bun.IDB {
	if tx, ok := TxFromContext(ctx); ok {
		return tx
	}
	return d.Slave()
}

func (d *DB) Writer(ctx context.Context) bun.IDB {
	if tx, ok := TxFromContext(ctx); ok {
		return tx
	}
	return d.master
}
