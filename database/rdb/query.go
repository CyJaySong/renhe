package rdb

import (
	"context"
	"database/sql"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

// --- Read operations → Reader (slave, tx fallback) ---

// NewSelect 创建 SELECT 查询，自动路由到从库或事务。
func (d *DB) NewSelect(ctx context.Context) *bun.SelectQuery {
	return d.Reader(ctx).NewSelect()
}

// QueryContext 执行原生查询，自动路由到从库或事务。
func (d *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return d.Reader(ctx).QueryContext(ctx, query, args...)
}

// QueryRowContext 执行原生单行查询，自动路由到从库或事务。
func (d *DB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return d.Reader(ctx).QueryRowContext(ctx, query, args...)
}

// --- Write operations → Writer (master, tx fallback) ---

// NewInsert 创建 INSERT 查询，自动路由到主库或事务。
func (d *DB) NewInsert(ctx context.Context) *bun.InsertQuery {
	return d.Writer(ctx).NewInsert()
}

// NewUpdate 创建 UPDATE 查询，自动路由到主库或事务。
func (d *DB) NewUpdate(ctx context.Context) *bun.UpdateQuery {
	return d.Writer(ctx).NewUpdate()
}

// NewDelete 创建 DELETE 查询，自动路由到主库或事务。
func (d *DB) NewDelete(ctx context.Context) *bun.DeleteQuery {
	return d.Writer(ctx).NewDelete()
}

// NewMerge 创建 MERGE 查询，自动路由到主库或事务。
func (d *DB) NewMerge(ctx context.Context) *bun.MergeQuery {
	return d.Writer(ctx).NewMerge()
}

// NewRaw 创建原生 SQL 查询，自动路由到主库或事务。
func (d *DB) NewRaw(ctx context.Context, query string, args ...any) *bun.RawQuery {
	return d.Writer(ctx).NewRaw(query, args...)
}

// ExecContext 执行原生写操作，自动路由到主库或事务。
func (d *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return d.Writer(ctx).ExecContext(ctx, query, args...)
}

// --- DDL operations → always master ---

// Dialect 返回主库的 SQL 方言。
func (d *DB) Dialect() schema.Dialect {
	return d.master.Dialect()
}

// NewValues 创建 VALUES 子查询，始终使用主库。
func (d *DB) NewValues(model any) *bun.ValuesQuery {
	return d.master.NewValues(model)
}

// NewCreateTable 创建建表查询，始终路由到主库。
func (d *DB) NewCreateTable(ctx context.Context) *bun.CreateTableQuery {
	return d.Writer(ctx).NewCreateTable()
}

// NewDropTable 创建删表查询，始终路由到主库。
func (d *DB) NewDropTable(ctx context.Context) *bun.DropTableQuery {
	return d.Writer(ctx).NewDropTable()
}

// NewCreateIndex 创建索引查询，始终路由到主库。
func (d *DB) NewCreateIndex(ctx context.Context) *bun.CreateIndexQuery {
	return d.Writer(ctx).NewCreateIndex()
}

// NewDropIndex 创建删索引查询，始终路由到主库。
func (d *DB) NewDropIndex(ctx context.Context) *bun.DropIndexQuery {
	return d.Writer(ctx).NewDropIndex()
}

// NewTruncateTable 创建清空表查询，始终路由到主库。
func (d *DB) NewTruncateTable(ctx context.Context) *bun.TruncateTableQuery {
	return d.Writer(ctx).NewTruncateTable()
}

// NewAddColumn 创建加列查询，始终路由到主库。
func (d *DB) NewAddColumn(ctx context.Context) *bun.AddColumnQuery {
	return d.Writer(ctx).NewAddColumn()
}

// NewDropColumn 创建删列查询，始终路由到主库。
func (d *DB) NewDropColumn(ctx context.Context) *bun.DropColumnQuery {
	return d.Writer(ctx).NewDropColumn()
}

// --- Transaction operations → always master ---

// BeginTx 在主库上开启事务。
func (d *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (bun.Tx, error) {
	return d.Writer(ctx).BeginTx(ctx, opts)
}

// BeginTxWithCtx 开启事务并将其注入返回的 context，后续操作自动路由到事务。
func (d *DB) BeginTxWithCtx(ctx context.Context, opts *sql.TxOptions) (bun.Tx, context.Context, error) {
	tx, err := d.Writer(ctx).BeginTx(ctx, opts)
	if err != nil {
		return tx, ctx, err
	}
	return tx, WithTx(ctx, tx), nil
}

// RunInTx 在事务中执行 fn，fn 收到的 ctx 已注入事务。fn 返回 error 时自动回滚，否则提交。
func (d *DB) RunInTx(ctx context.Context, opts *sql.TxOptions, fn func(ctx context.Context, tx bun.Tx) error) error {
	return d.master.RunInTx(ctx, opts, func(ctx context.Context, tx bun.Tx) error {
		return fn(WithTx(ctx, tx), tx)
	})
}
