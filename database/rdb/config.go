package rdb

import (
	"fmt"
	"github.com/cyjaysong/renhe/os/rcfg"
	"time"
)

// 连接池配置
type poolConfig struct {
	MaxOpenConns    int           `yaml:"maxOpenConns"`
	MaxIdleConns    int           `yaml:"maxIdleConns"`
	ConnMaxLifetime time.Duration `yaml:"connMaxLifetime"`
	ConnMaxIdleTime time.Duration `yaml:"connMaxIdleTime"`
}

type Config struct {
	Pool     poolConfig `yaml:",squash"`
	DSN      string     `yaml:"dsn"`
	SlaveDSN []string   `yaml:"slave"`
}

func loadConfig(name string) (cfg Config, err error) {
	key := "database." + name
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
	if cfg.DSN == "" {
		return Config{}, fmt.Errorf("config `%s`: dsn is required", key)
	}
	return
}
