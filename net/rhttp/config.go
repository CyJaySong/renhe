// Package rhttp 提供 HTTP 服务器封装，基于 echo 框架，集成配置加载、验证器和日志。
package rhttp

import (
	"fmt"

	"github.com/cyjaysong/renhe/os/rcfg"
)

// Config HTTP 服务器配置。
type Config struct {
	Address string `yaml:"address"`
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
	return
}
