package redis

import (
	"context"
	"fmt"
	"github.com/cyjaysong/renhe/os/rctx"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// Redis 封装 go-redis UniversalClient，支持单机和集群模式。
// 地址列表超过 1 个时自动启用集群模式。
type Redis struct {
	cfg  Config
	name string
	goredis.UniversalClient
}

// newRedis 根据配置创建 Redis 客户端，连接成功后执行 Ping 验证。
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
	ctx, cancel := context.WithTimeout(rctx.GetInitCtx(), time.Second*3)
	err := client.Ping(ctx).Err()
	cancel()
	if err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis ping failed: %w", err)
	}
	return &Redis{cfg: cfg, name: name, UniversalClient: client}, nil
}

// Cfg 返回当前实例的配置。
func (r *Redis) Cfg() Config {
	return r.cfg
}

// Name 返回实例名称（如 "default"）。
func (r *Redis) Name() string {
	return r.name
}
