package rdb

import (
	"fmt"
	"time"

	"github.com/cyjaysong/renhe/os/rcfg"
)

// 默认值
const (
	defaultSlowQueryThreshold  = 5 * time.Second
	defaultHealthCheckInterval = 5 * time.Second
	defaultPingTimeout         = 3 * time.Second
)

// poolConfig 数据库连接池配置，通过 yaml squash 嵌入 Config。
type poolConfig struct {
	MaxOpenConns    int           `yaml:"maxOpenConns"`
	MaxIdleConns    int           `yaml:"maxIdleConns"`
	ConnMaxLifetime time.Duration `yaml:"connMaxLifetime"`
	ConnMaxIdleTime time.Duration `yaml:"connMaxIdleTime"`
}

// Config 数据库配置，包含主库 DSN、从库 DSN 列表和连接池参数。
type Config struct {
	Pool     poolConfig `yaml:",squash"`
	DSN      string     `yaml:"dsn"`
	SlaveDSN []string   `yaml:"slave"`
	// SlowQueryThreshold 慢查询阈值；0 表示默认 5s；负数表示关闭慢查询标记。
	SlowQueryThreshold time.Duration `yaml:"slowQueryThreshold"`
	// HealthCheckInterval 主从健康检查间隔；0 表示默认 5s。
	HealthCheckInterval time.Duration `yaml:"healthCheckInterval"`
	// PingTimeout 启动软 Ping 与健康检查 Ping 超时；0 表示默认 3s。
	PingTimeout time.Duration `yaml:"pingTimeout"`
}

func (c Config) slowQueryThreshold() time.Duration {
	if c.SlowQueryThreshold < 0 {
		return 0
	}
	if c.SlowQueryThreshold == 0 {
		return defaultSlowQueryThreshold
	}
	return c.SlowQueryThreshold
}

func (c Config) healthCheckInterval() time.Duration {
	if c.HealthCheckInterval <= 0 {
		return defaultHealthCheckInterval
	}
	return c.HealthCheckInterval
}

func (c Config) pingTimeout() time.Duration {
	if c.PingTimeout <= 0 {
		return defaultPingTimeout
	}
	return c.PingTimeout
}

// loadConfig 从全局配置中读取 database.<name> 下的数据库配置并反序列化。
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
