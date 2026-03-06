package rdb

import (
	"context"
	"database/sql"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

// --- Read operations → Reader (slave, tx fallback) ---

func (d *DB) NewSelect(ctx context.Context) *bun.SelectQuery {
	return d.Reader(ctx).NewSelect()
}

func (d *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return d.Reader(ctx).QueryContext(ctx, query, args...)
}

func (d *DB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return d.Reader(ctx).QueryRowContext(ctx, query, args...)
}

// --- Write operations → Writer (master, tx fallback) ---

func (d *DB) NewInsert(ctx context.Context) *bun.InsertQuery {
	return d.Writer(ctx).NewInsert()
}

func (d *DB) NewUpdate(ctx context.Context) *bun.UpdateQuery {
	return d.Writer(ctx).NewUpdate()
}

func (d *DB) NewDelete(ctx context.Context) *bun.DeleteQuery {
	return d.Writer(ctx).NewDelete()
}

func (d *DB) NewMerge(ctx context.Context) *bun.MergeQuery {
	return d.Writer(ctx).NewMerge()
}

func (d *DB) NewRaw(ctx context.Context, query string, args ...any) *bun.RawQuery {
	return d.Writer(ctx).NewRaw(query, args...)
}

func (d *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return d.Writer(ctx).ExecContext(ctx, query, args...)
}

// --- DDL operations → always master ---

func (d *DB) Dialect() schema.Dialect {
	return d.master.Dialect()
}

func (d *DB) NewValues(model any) *bun.ValuesQuery {
	return d.master.NewValues(model)
}

func (d *DB) NewCreateTable(ctx context.Context) *bun.CreateTableQuery {
	return d.Writer(ctx).NewCreateTable()
}

func (d *DB) NewDropTable(ctx context.Context) *bun.DropTableQuery {
	return d.Writer(ctx).NewDropTable()
}

func (d *DB) NewCreateIndex(ctx context.Context) *bun.CreateIndexQuery {
	return d.Writer(ctx).NewCreateIndex()
}

func (d *DB) NewDropIndex(ctx context.Context) *bun.DropIndexQuery {
	return d.Writer(ctx).NewDropIndex()
}

func (d *DB) NewTruncateTable(ctx context.Context) *bun.TruncateTableQuery {
	return d.Writer(ctx).NewTruncateTable()
}

func (d *DB) NewAddColumn(ctx context.Context) *bun.AddColumnQuery {
	return d.Writer(ctx).NewAddColumn()
}

func (d *DB) NewDropColumn(ctx context.Context) *bun.DropColumnQuery {
	return d.Writer(ctx).NewDropColumn()
}

// --- Transaction operations → always master ---

func (d *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (bun.Tx, error) {
	return d.Writer(ctx).BeginTx(ctx, opts)
}

func (d *DB) BeginTxWithCtx(ctx context.Context, opts *sql.TxOptions) (bun.Tx, context.Context, error) {
	tx, err := d.Writer(ctx).BeginTx(ctx, opts)
	if err != nil {
		return tx, ctx, err
	}
	return tx, WithTx(ctx, tx), nil
}

func (d *DB) RunInTx(ctx context.Context, opts *sql.TxOptions, fn func(ctx context.Context) error) error {
	return d.master.RunInTx(ctx, opts, func(ctx context.Context, tx bun.Tx) error {
		return fn(WithTx(ctx, tx))
	})
}
