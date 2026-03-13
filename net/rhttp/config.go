// Package rhttp 提供 HTTP 服务器封装，基于 echo 框架，集成配置加载、验证器和日志。
package rhttp

import (
	"github.com/cyjaysong/renhe/os/rcfg"
	"github.com/cyjaysong/renhe/os/rctx"
	"github.com/cyjaysong/renhe/os/rlog"
)

// Config HTTP 服务器配置。
type Config struct {
	Address string `yaml:"address"`
}

// loadConfig 从全局配置 httpSrv 下读取 HTTP 服务器配置并反序列化。
func loadConfig() (cfg Config) {
	key := "httpSrv"
	allCfg := rcfg.Cfg()
	if !allCfg.IsSet(key) {
		rlog.Log().Warnf(rctx.GetInitCtx(), "config `%s` not found", key)
		return
	}
	if sub := allCfg.Sub(key); sub == nil {
		rlog.Log().Warnf(rctx.GetInitCtx(), "config `%s` is empty", key)
		return
	} else if err := sub.Unmarshal(&cfg, rcfg.YamlTagOption); err != nil {
		rlog.Log().Warnf(rctx.GetInitCtx(), "config `%s`: unmarshal failed: %s", key, err)
		return
	}
	if cfg.Address == "" {
		rlog.Log().Warnf(rctx.GetInitCtx(), "config `%s`: address is required", key)
	}
	return
}
