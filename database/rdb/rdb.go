package rdb

import (
	"context"
	"fmt"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/cyjaysong/renhe/os/rlog"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

// slaveNode 封装一个从库节点，包含连接和健康状态。
type slaveNode struct {
	dsn     string
	db      *bun.DB
	healthy atomic.Bool
}

// maskDSN 对 DSN 中的密码进行脱敏，替换为 "***"。
func maskDSN(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return "***"
	}
	if u.User != nil {
		u.User = url.UserPassword(u.User.Username(), "***")
	}
	return u.String()
}

// DB 封装主从数据库实例，提供读写分离和自动故障转移能力。
type DB struct {
	cfg    Config
	name   string
	master *bun.DB
	slaves []*slaveNode
}

const healthCheckInterval = 5 * time.Second

// newDB 根据配置创建主库和从库实例。若存在从库，启动后台健康检查 goroutine。
func newDB(name string, cfg Config) (*DB, error) {
	bunQueryHookForLog := rlog.Log().BunQueryHook(5 * time.Second)

	sqlDb, err := openSqlDB(cfg.DSN, cfg.Pool)
	if err != nil {
		return nil, fmt.Errorf("rdb: failed to open master: %w", err)
	}

	masterBun := bun.NewDB(sqlDb, pgdialect.New())
	masterBun = masterBun.WithQueryHook(bunQueryHookForLog)
	d := &DB{cfg: cfg, name: name, master: masterBun}

	for i, dsn := range cfg.SlaveDSN {
		if sqlDb, err = openSqlDB(dsn, cfg.Pool); err != nil {
			return nil, fmt.Errorf("rdb: failed to open slave[%d]: %w", i, err)
		}
		slaveBun := bun.NewDB(sqlDb, pgdialect.New())
		slaveBun = slaveBun.WithQueryHook(bunQueryHookForLog)
		node := &slaveNode{dsn: dsn, db: slaveBun}
		node.healthy.Store(true)
		d.slaves = append(d.slaves, node)
	}

	if len(d.slaves) > 0 {
		go d.healthCheck()
	}
	return d, nil
}

// Cfg 返回当前实例的配置。
func (d *DB) Cfg() Config {
	return d.cfg
}

// Name 返回实例名称（如 "default"）。
func (d *DB) Name() string {
	return d.name
}

// Master 返回主库 bun.DB 实例。
func (d *DB) Master() *bun.DB {
	return d.master
}

// Slave 返回一个健康的从库实例（最小活跃连接数策略）。
// 若所有从库不健康或无从库配置，回退到主库。
func (d *DB) Slave() *bun.DB {
	if len(d.slaves) == 0 {
		return d.master
	}
	var best *slaveNode
	bestActive := int(^uint(0) >> 1)
	for _, s := range d.slaves {
		if !s.healthy.Load() {
			continue
		}
		stats := s.db.DB.Stats()
		active := stats.InUse
		if active < bestActive {
			bestActive = active
			best = s
		}
	}
	if best == nil {
		return d.master
	}
	return best.db
}

// healthCheck 后台定期 Ping 所有从库节点，标记健康状态。
// 节点宕机或恢复时记录日志，对业务层完全透明。
func (d *DB) healthCheck() {
	ticker := time.NewTicker(healthCheckInterval)
	defer ticker.Stop()
	for range ticker.C {
		for _, s := range d.slaves {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			err := s.db.DB.PingContext(ctx)
			cancel()
			was := s.healthy.Load()
			now := err == nil
			s.healthy.Store(now)
			if was && !now {
				rlog.Log().Warnf(context.Background(), "rdb[%s]: slave %s marked unhealthy: %v", d.name, maskDSN(s.dsn), err)
			} else if !was && now {
				rlog.Log().Infof(context.Background(), "rdb[%s]: slave %s recovered", d.name, maskDSN(s.dsn))
			}
		}
	}
}
