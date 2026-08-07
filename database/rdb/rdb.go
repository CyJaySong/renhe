package rdb

import (
	"context"
	"fmt"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cyjaysong/renhe/os/rctx"
	"github.com/cyjaysong/renhe/os/rlog"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/extra/bunotel"
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
		return raw
	}
	if u.User != nil {
		if _, has := u.User.Password(); has {
			u.User = url.UserPassword(u.User.Username(), "***")
		}
	}
	return u.String()
}

// DB 封装主从数据库实例，提供读写分离和自动故障转移能力。
type DB struct {
	cfg           Config
	name          string
	master        *bun.DB
	slaves        []*slaveNode
	rrIdx         atomic.Uint64
	masterHealthy atomic.Bool
	stopHC        chan struct{}
	closeOnce     sync.Once
}

// newDB 根据配置创建主库和从库实例，并启动后台健康检查（主库 + 从库）。
// Open 失败才返回 error；Ping 失败仅告警，实例仍保留。
func newDB(name string, cfg Config) (*DB, error) {
	bunQueryHookForLog := rlog.Log().BunQueryHook(cfg.slowQueryThreshold())
	bunOtelHook := bunotel.NewQueryHook(bunotel.WithDBName(name))
	pingTO := cfg.pingTimeout()

	sqlDb, err := openSqlDB(cfg.DSN, cfg.Pool, name+"/master", pingTO)
	if err != nil {
		return nil, fmt.Errorf("rdb: failed to open master: %w", err)
	}

	masterBun := bun.NewDB(sqlDb, pgdialect.New())
	masterBun = masterBun.WithQueryHook(bunOtelHook)
	masterBun = masterBun.WithQueryHook(bunQueryHookForLog)
	d := &DB{
		cfg:    cfg,
		name:   name,
		master: masterBun,
		stopHC: make(chan struct{}),
	}
	d.masterHealthy.Store(true)

	for i, dsn := range cfg.SlaveDSN {
		if sqlDb, err = openSqlDB(dsn, cfg.Pool, fmt.Sprintf("%s/slave[%d]", name, i), pingTO); err != nil {
			_ = d.Close()
			return nil, fmt.Errorf("rdb: failed to open slave[%d]: %w", i, err)
		}
		slaveBun := bun.NewDB(sqlDb, pgdialect.New())
		slaveBun = slaveBun.WithQueryHook(bunOtelHook)
		slaveBun = slaveBun.WithQueryHook(bunQueryHookForLog)
		node := &slaveNode{dsn: dsn, db: slaveBun}
		node.healthy.Store(true)
		d.slaves = append(d.slaves, node)
	}

	go d.healthCheck()
	return d, nil
}

// Cfg 返回当前实例的配置。
func (d *DB) Cfg() Config { return d.cfg }

// Name 返回实例名称（如 "default"）。
func (d *DB) Name() string { return d.name }

// Master 返回主库 bun.DB 实例。
func (d *DB) Master() *bun.DB { return d.master }

// MasterHealthy 返回主库最近一次健康检查是否成功（可观测，不阻断查询）。
func (d *DB) MasterHealthy() bool { return d.masterHealthy.Load() }

// Slave 返回一个健康的从库实例（Round-Robin）；无从库或不健康时回退主库。
func (d *DB) Slave() *bun.DB {
	n := len(d.slaves)
	if n == 0 {
		return d.master
	}
	idx := d.rrIdx.Add(1)
	for i := range n {
		s := d.slaves[(int(idx)+i)%n]
		if s.healthy.Load() {
			return s.db
		}
	}
	return d.master
}

// Close 停止健康检查并关闭主从连接。
func (d *DB) Close() error {
	var first error
	d.closeOnce.Do(func() {
		if d.stopHC != nil {
			close(d.stopHC)
		}
		for _, s := range d.slaves {
			if s.db != nil {
				if err := s.db.Close(); err != nil && first == nil {
					first = err
				}
			}
		}
		if d.master != nil {
			if err := d.master.Close(); err != nil && first == nil {
				first = err
			}
		}
	})
	return first
}

// healthCheck 后台定期 Ping 主库与从库，仅打日志/更新状态，不剔除主库实例。
func (d *DB) healthCheck() {
	interval := d.cfg.healthCheckInterval()
	pingTO := d.cfg.pingTimeout()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-d.stopHC:
			return
		case <-ticker.C:
			// 主库
			ctx, cancel := context.WithTimeout(rctx.GetInitCtx(), pingTO)
			err := d.master.DB.PingContext(ctx)
			cancel()
			was := d.masterHealthy.Load()
			now := err == nil
			d.masterHealthy.Store(now)
			if was && !now {
				rlog.Log().Warn(ctx, "rdb: master marked unhealthy",
					"name", d.name, "dsn", maskDSN(d.cfg.DSN), "err", err)
			} else if !was && now {
				rlog.Log().Info(ctx, "rdb: master recovered",
					"name", d.name, "dsn", maskDSN(d.cfg.DSN))
			}
			// 从库
			for _, s := range d.slaves {
				ctx, cancel := context.WithTimeout(rctx.GetInitCtx(), pingTO)
				err := s.db.DB.PingContext(ctx)
				cancel()
				was := s.healthy.Load()
				now := err == nil
				s.healthy.Store(now)
				if was && !now {
					rlog.Log().Warn(ctx, "rdb: slave marked unhealthy",
						"name", d.name, "dsn", maskDSN(s.dsn), "err", err)
				} else if !was && now {
					rlog.Log().Info(ctx, "rdb: slave recovered",
						"name", d.name, "dsn", maskDSN(s.dsn))
				}
			}
		}
	}
}
