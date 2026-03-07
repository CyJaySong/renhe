package rdb

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// openSqlDB 使用 pgx 驱动打开 SQL 连接并应用连接池配置。
func openSqlDB(dsn string, p poolConfig) (sqlDb *sql.DB, err error) {
	if sqlDb, err = sql.Open("pgx", dsn); err != nil {
		return
	}
	if p.MaxOpenConns > 0 {
		sqlDb.SetMaxOpenConns(p.MaxOpenConns)
	}
	if p.MaxIdleConns > 0 {
		sqlDb.SetMaxIdleConns(p.MaxIdleConns)
	}
	if p.ConnMaxLifetime > 0 {
		sqlDb.SetConnMaxLifetime(p.ConnMaxLifetime)
	}
	if p.ConnMaxIdleTime > 0 {
		sqlDb.SetConnMaxIdleTime(p.ConnMaxIdleTime)
	}
	return
}
