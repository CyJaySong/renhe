package rdb

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/cyjaysong/renhe/os/rcfg"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
)

var (
	instances = make(map[string]*DB)
	mu        sync.RWMutex
)

type DB struct {
	name   string
	master *bun.DB
	slaves []*bun.DB
}

type PoolConfig struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type Config struct {
	DSN      string
	Pool     PoolConfig
	SlaveDSN []string
}

func applyPool(db *sql.DB, p PoolConfig) {
	if p.MaxOpenConns > 0 {
		db.SetMaxOpenConns(p.MaxOpenConns)
	}
	if p.MaxIdleConns > 0 {
		db.SetMaxIdleConns(p.MaxIdleConns)
	}
	if p.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(p.ConnMaxLifetime)
	}
	if p.ConnMaxIdleTime > 0 {
		db.SetConnMaxIdleTime(p.ConnMaxIdleTime)
	}
}

func newDB(cfg Config) (*DB, error) {
	sqlDb, err := openSqlDB(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("rdb: failed to open master: %w", err)
	}
	applyPool(sqlDb, cfg.Pool)

	d := &DB{master: bun.NewDB(sqlDb, pgdialect.New())}

	for i, dsn := range cfg.SlaveDSN {
		if sqlDb, err = openSqlDB(dsn); err != nil {
			return nil, fmt.Errorf("rdb: failed to open slave[%d]: %w", i, err)
		}
		applyPool(sqlDb, cfg.Pool)
		d.slaves = append(d.slaves, bun.NewDB(sqlDb, pgdialect.New()))
	}

	return d, nil
}

func Instance(name ...string) *DB {
	n := "default"
	if len(name) > 0 && name[0] != "" {
		n = name[0]
	}

	mu.RLock()
	if d, ok := instances[n]; ok {
		mu.RUnlock()
		return d
	}
	mu.RUnlock()

	mu.Lock()
	defer mu.Unlock()
	if d, ok := instances[n]; ok {
		return d
	}
	cfg, err := loadConfig(n)
	if err != nil {
		fmt.Printf("rdb: %v\n", err)
		return nil
	}
	d, err := newDB(cfg)
	if err != nil {
		fmt.Printf("rdb: failed to create instance %q: %v\n", n, err)
		return nil
	}
	d.name = n
	instances[n] = d
	return d
}

func loadConfig(name string) (Config, error) {
	v := rcfg.Instance()
	key := "database." + name
	if !v.IsSet(key) {
		return Config{}, fmt.Errorf("database config %q not found", name)
	}
	sub := v.Sub(key)
	if sub == nil {
		return Config{}, fmt.Errorf("database config %q is empty", name)
	}
	cfg := Config{
		DSN:      sub.GetString("dsn"),
		SlaveDSN: sub.GetStringSlice("slave"),
		Pool: PoolConfig{
			MaxOpenConns:    sub.GetInt("maxOpenConns"),
			MaxIdleConns:    sub.GetInt("maxIdleConns"),
			ConnMaxLifetime: sub.GetDuration("connMaxLifetime"),
			ConnMaxIdleTime: sub.GetDuration("connMaxIdleTime"),
		},
	}
	if cfg.DSN == "" {
		return Config{}, fmt.Errorf("database config %q: dsn is required", name)
	}
	return cfg, nil
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

func (d *DB) Close() error {
	var firstErr error
	if d.master != nil {
		if err := d.master.Close(); err != nil {
			firstErr = err
		}
	}
	for _, s := range d.slaves {
		if err := s.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
