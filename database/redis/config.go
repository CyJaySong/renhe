// Package redis 提供 Redis 客户端封装，支持单机和集群模式，基于 go-redis/v9 UniversalClient 实现。
package redis

import (
	"fmt"
	"time"

	"github.com/cyjaysong/renhe/os/rcfg"
)

// Config Redis 连接配置。
type Config struct {
	// 地址列表: 单机模式填1个, 集群模式填多个种子节点
	Address  []string `yaml:"address"`
	DB       int      `yaml:"db"`       // 数据库索引(仅单机模式)
	Username string   `yaml:"username"` // 访问授权用户
	Password string   `yaml:"password"` // 访问授权密码
	// 连接池
	MinIdleConns    int           `yaml:"minIdleConns"`    // 最小空闲连接数
	MaxIdleConns    int           `yaml:"maxIdleConns"`    // 最大空闲连接数
	PoolSize        int           `yaml:"poolSize"`        // 连接池大小
	ConnMaxIdleTime time.Duration `yaml:"connMaxIdleTime"` // 连接最大空闲时间
	ConnMaxLifetime time.Duration `yaml:"connMaxLifetime"` // 连接最长存活时间
	PoolTimeout     time.Duration `yaml:"poolTimeout"`     // 等待连接池连接的超时时间
	// 超时
	DialTimeout  time.Duration `yaml:"dialTimeout"`  // TCP连接超时
	ReadTimeout  time.Duration `yaml:"readTimeout"`  // TCP读超时
	WriteTimeout time.Duration `yaml:"writeTimeout"` // TCP写超时
}

// loadConfig 从全局配置中读取 redis.<name> 下的 Redis 配置并反序列化。
func loadConfig(name string) (cfg Config, err error) {
	key := "redis." + name
	allCfg := rcfg.Cfg()
	if !allCfg.IsSet(key) {
		return Config{}, fmt.Errorf("config `%s` not found", key)
	}
	sub := allCfg.Sub(key)
	if sub == nil {
		return Config{}, fmt.Errorf("config `%s` is empty", key)
	}
	if err = sub.Unmarshal(&cfg, rcfg.YamlTagOption); err != nil {
		return Config{}, fmt.Errorf("config `%s`: unmarshal failed: %w", key, err)
	}
	if len(cfg.Address) == 0 {
		return Config{}, fmt.Errorf("config `%s`: address is required", key)
	}
	return
}
