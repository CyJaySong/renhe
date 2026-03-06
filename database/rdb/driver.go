package rdb

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func openSqlDB(dsn string) (*sql.DB, error) {
	return sql.Open("pgx", dsn)
}
