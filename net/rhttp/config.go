// Package rhttp 提供 HTTP 服务器封装，基于 echo 框架，集成配置加载、验证器和日志。
package rhttp

import (
	"fmt"
	"time"

	"github.com/cyjaysong/renhe/os/rcfg"
)

// 默认优雅退出等待时间（与 echo.StartConfig 默认一致）
const defaultGracefulTimeout = 10 * time.Second

// Config HTTP 服务器配置。
type Config struct {
	Address string `yaml:"address"`
	// GracefulTimeout SIGINT/SIGTERM 后等待进行中请求结束的时间；0 表示默认 10s。
	GracefulTimeout time.Duration `yaml:"gracefulTimeout"`
}

// gracefulTimeout 解析有效优雅退出超时。
func (c Config) gracefulTimeout() time.Duration {
	if c.GracefulTimeout <= 0 {
		return defaultGracefulTimeout
	}
	return c.GracefulTimeout
}

// loadConfig 从全局配置 httpSrv 下读取 HTTP 服务器配置并反序列化。
func loadConfig() (cfg Config, err error) {
	key := "httpSrv"
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
	if cfg.Address == "" {
		return Config{}, fmt.Errorf("config `%s`: address is required", key)
	}
	return cfg, nil
}
