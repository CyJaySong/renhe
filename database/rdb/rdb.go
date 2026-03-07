package rdb

import (
	"fmt"
	"time"

	"github.com/cyjaysong/renhe/os/rlog"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

type DB struct {
	cfg    Config
	name   string
	master *bun.DB
	slaves []*bun.DB
}

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
		d.slaves = append(d.slaves, slaveBun)
	}
	return d, nil
}

func (d *DB) Cfg() Config {
	return d.cfg
}

func (d *DB) Name() string {
	return d.name
}

func (d *DB) Master() *bun.DB {
	return d.master
}

func (d *DB) Slave() *bun.DB {
	if len(d.slaves) == 0 {
		return d.master
	}
	var best *bun.DB
	bestActive := int(^uint(0) >> 1)
	for _, s := range d.slaves {
		stats := s.DB.Stats()
		active := stats.InUse
		if active < bestActive {
			bestActive = active
			best = s
		}
	}
	return best
}
