package rdb

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestMaskDSN(t *testing.T) {
	in := "postgres://user:secret@127.0.0.1:5432/db?sslmode=disable"
	out := maskDSN(in)
	if strings.Contains(out, "secret") {
		t.Fatalf("password leaked: %s", out)
	}
	if !strings.Contains(out, "user") || !strings.Contains(out, "127.0.0.1") {
		t.Fatalf("unexpected: %s", out)
	}
}

func TestSoftPing_Unreachable_DoesNotClose(t *testing.T) {
	// Open 通常成功；Ping 失败时连接仍可用（未 Close）
	db, err := sql.Open("pgx", "postgres://u:p@127.0.0.1:1/db?sslmode=disable&connect_timeout=1")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	pingErr := softPing(ctx, db)
	if pingErr == nil {
		t.Fatal("expected ping error on unreachable port")
	}
	// 软失败：库对象仍在，可再次 Close（若已被关会返回 err）
	if err := db.PingContext(context.Background()); err == nil {
		// 极少数环境若 1 端口可达则跳过
		t.Skip("port 1 unexpectedly reachable")
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close after soft ping: %v", err)
	}
}

func TestOpenSqlDB_SoftPingKeepsInstance(t *testing.T) {
	// 替换告警，避免 rcfg 未初始化导致 Fatal
	var warned bool
	old := pingWarn
	pingWarn = func(ctx context.Context, logLabel, dsn string, pingErr error) {
		warned = true
	}
	defer func() { pingWarn = old }()

	db, err := openSqlDB(
		"postgres://u:p@127.0.0.1:1/db?sslmode=disable&connect_timeout=1",
		poolConfig{},
		"test/master",
		500*time.Millisecond,
	)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if db == nil {
		t.Fatal("db nil")
	}
	if !warned {
		t.Fatal("expected pingWarn on soft fail")
	}
	_ = db.Close()
}
