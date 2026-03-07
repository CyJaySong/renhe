package redis

import (
	"fmt"
	"time"

	"github.com/cyjaysong/renhe/os/rcfg"
)

type Config struct {
	Addr            string        `yaml:"addr"`            // Redis地址
	DB              int           `yaml:"db"`              // 数据库索引
	Username        string        `yaml:"username"`        // 访问授权用户
	Password        string        `yaml:"password"`        // 访问授权密码
	MinIdle         int           `yaml:"minIdle"`         // 允许闲置的最小连接数
	MaxIdle         int           `yaml:"maxIdle"`         // 允许闲置的最大连接数(0表示不限制)
	MaxActive       int           `yaml:"maxActive"`       // 最大连接数量限制(0表示不限制)
	IdleTimeout     time.Duration `yaml:"idleTimeout"`     // 连接最大空闲时间
	MaxConnLifetime time.Duration `yaml:"maxConnLifetime"` // 连接最长存活时间
	DialTimeout     time.Duration `yaml:"dialTimeout"`     // TCP连接的超时时间
	ReadTimeout     time.Duration `yaml:"readTimeout"`     // TCP的Read操作超时时间
	WriteTimeout    time.Duration `yaml:"writeTimeout"`    // TCP的Write操作超时时间
	WaitTimeout     time.Duration `yaml:"waitTimeout"`     // 等待连接池连接的超时时间
}

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
	if cfg.Addr == "" {
		return Config{}, fmt.Errorf("config `%s`: address is required", key)
	}
	return
}
