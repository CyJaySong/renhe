package rdb

import (
	"context"
	"database/sql"
	"time"

	"github.com/cyjaysong/renhe/os/rctx"
	"github.com/cyjaysong/renhe/os/rlog"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// openSqlDB 使用 pgx 打开连接并应用连接池。
// Open 失败才返回 error；Ping 软失败只告警，仍返回 *sql.DB。
func openSqlDB(dsn string, p poolConfig, logLabel string, pingTimeout time.Duration) (sqlDb *sql.DB, err error) {
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
	if pingTimeout <= 0 {
		pingTimeout = defaultPingTimeout
	}
	// 尽力 Ping：失败不 Close、不阻断创建
	ctx, cancel := context.WithTimeout(rctx.GetInitCtx(), pingTimeout)
	defer cancel()
	if pingErr := softPing(ctx, sqlDb); pingErr != nil {
		// 日志失败不应用影响返回（测路径可替换 pingWarn）
		pingWarn(ctx, logLabel, dsn, pingErr)
	}
	return sqlDb, nil
}

// softPing 仅执行 Ping，不关闭连接。
func softPing(ctx context.Context, sqlDb *sql.DB) error {
	return sqlDb.PingContext(ctx)
}

// pingWarn 启动软 Ping 失败时的告警（测试可替换）。
var pingWarn = func(ctx context.Context, logLabel, dsn string, pingErr error) {
	rlog.Log().Warn(ctx, "rdb: ping failed, keep instance for auto-reconnect",
		"target", logLabel, "dsn", maskDSN(dsn), "err", pingErr)
}
