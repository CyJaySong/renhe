package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

type Redis struct {
	cfg  Config
	name string
	*goredis.Client
}

func newRedis(name string, cfg Config) (*Redis, error) {
	opts := &goredis.Options{
		Addr:            cfg.Addr,
		DB:              cfg.DB,
		Username:        cfg.Username,
		Password:        cfg.Password,
		MinIdleConns:    cfg.MinIdle,
		MaxIdleConns:    cfg.MaxIdle,
		PoolSize:        cfg.MaxActive,
		ConnMaxIdleTime: cfg.IdleTimeout,
		ConnMaxLifetime: cfg.MaxConnLifetime,
		DialTimeout:     cfg.DialTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		PoolTimeout:     cfg.WaitTimeout,
	}

	client := goredis.NewClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}
	return &Redis{cfg: cfg, name: name, Client: client}, nil
}

func (r *Redis) Cfg() Config {
	return r.cfg
}

func (r *Redis) Name() string {
	return r.name
}
