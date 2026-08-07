package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/cyjaysong/renhe/os/rctx"
	"github.com/cyjaysong/renhe/os/rlog"
	"github.com/redis/go-redis/extra/redisotel/v9"
	goredis "github.com/redis/go-redis/v9"
)

// Redis 封装 go-redis UniversalClient，支持单机和集群模式。
// 模式由 Config.Mode 控制（standalone / cluster / auto）。
type Redis struct {
	cfg  Config
	name string
	goredis.UniversalClient
}

// newRedis 根据配置创建 Redis 客户端。
// 客户端创建 / OTel 注入失败才返回 error；Ping 为尽力而为：失败只 Warn，仍返回实例，
// 便于 Redis 短暂不可达时保留连接池，恢复后自动重连（无需重启进程）。
func newRedis(name string, cfg Config) (*Redis, error) {
	cluster := cfg.isCluster()
	opts := &goredis.UniversalOptions{
		Addrs:           cfg.Address,
		IsClusterMode:   cluster,
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
	if err := redisotel.InstrumentTracing(client); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("redis otel tracing: %w", err)
	}
	// 尽力 Ping：失败不 Close、不阻断创建
	ctx, cancel := context.WithTimeout(rctx.GetInitCtx(), 3*time.Second)
	pingErr := client.Ping(ctx).Err()
	cancel()
	if pingErr != nil {
		mode := "standalone"
		if cluster {
			mode = "cluster"
		}
		rlog.Log().Warn(ctx, "redis: ping failed, keep instance for auto-reconnect",
			"name", name, "mode", mode, "address", cfg.Address, "err", pingErr)
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
