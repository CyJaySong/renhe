package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

type Redis struct {
	cfg  Config
	name string
	goredis.UniversalClient
}

func newRedis(name string, cfg Config) (*Redis, error) {
	opts := &goredis.UniversalOptions{
		Addrs:           cfg.Address,
		IsClusterMode:   len(cfg.Address) > 1,
		DB:              cfg.DB,
		Username:        cfg.Username,
		Password:        cfg.Password,
		MinIdleConns:    cfg.MinIdleConns,
		MaxIdleConns:    cfg.MaxIdleConns,
		PoolSize:        cfg.PoolSize,
		ConnMaxIdleTime: cfg.ConnMaxIdleTime,
		ConnMaxLifetime: cfg.ConnMaxLifetime,
		PoolTimeout:     cfg.PoolTimeout,
		DialTimeout:     cfg.DialTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
	}

	client := goredis.NewUniversalClient(opts)
	if err := client.Ping(context.Background()).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}
	return &Redis{cfg: cfg, name: name, UniversalClient: client}, nil
}

func (r *Redis) Cfg() Config {
	return r.cfg
}

func (r *Redis) Name() string {
	return r.name
}
