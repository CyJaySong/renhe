package rdb

import (
	"context"
	"database/sql"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

//var _ bun.IDB = (*DB)(nil)

func (d *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return d.bunDB.QueryContext(ctx, query, args...)
}

func (d *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return d.bunDB.ExecContext(ctx, query, args...)
}

func (d *DB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return d.bunDB.QueryRowContext(ctx, query, args...)
}

func (d *DB) Dialect() schema.Dialect {
	return d.bunDB.Dialect()
}

func (d *DB) NewValues(model any) *bun.ValuesQuery {
	return d.bunDB.NewValues(model)
}

func (d *DB) NewSelect() *bun.SelectQuery {
	return d.bunDB.NewSelect()
}

func (d *DB) NewInsert() *bun.InsertQuery {
	return d.bunDB.NewInsert()
}

func (d *DB) NewUpdate() *bun.UpdateQuery {
	return d.bunDB.NewUpdate()
}

func (d *DB) NewDelete() *bun.DeleteQuery {
	return d.bunDB.NewDelete()
}

func (d *DB) NewMerge() *bun.MergeQuery {
	return d.bunDB.NewMerge()
}

func (d *DB) NewRaw(query string, args ...any) *bun.RawQuery {
	return d.bunDB.NewRaw(query, args...)
}

func (d *DB) NewCreateTable() *bun.CreateTableQuery {
	return d.bunDB.NewCreateTable()
}

func (d *DB) NewDropTable() *bun.DropTableQuery {
	return d.bunDB.NewDropTable()
}

func (d *DB) NewCreateIndex() *bun.CreateIndexQuery {
	return d.bunDB.NewCreateIndex()
}

func (d *DB) NewDropIndex() *bun.DropIndexQuery {
	return d.bunDB.NewDropIndex()
}

func (d *DB) NewTruncateTable() *bun.TruncateTableQuery {
	return d.bunDB.NewTruncateTable()
}

func (d *DB) NewAddColumn() *bun.AddColumnQuery {
	return d.bunDB.NewAddColumn()
}

func (d *DB) NewDropColumn() *bun.DropColumnQuery {
	return d.bunDB.NewDropColumn()
}

func (d *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (tx Tx, err error) {
	tx.Tx, err = d.bunDB.BeginTx(ctx, opts)
	return
}

func (d *DB) RunInTx(ctx context.Context, opts *sql.TxOptions, f func(ctx context.Context, tx Tx) error) error {
	f2 := func(ctx context.Context, tx bun.Tx) error { return f(ctx, Tx{tx}) }
	return d.bunDB.RunInTx(ctx, opts, f2)
}
