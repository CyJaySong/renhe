package rhttp

import (
	"fmt"

	"github.com/cyjaysong/renhe/os/rcfg"
)

type Config struct {
	Address string `yaml:"address"`
}

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
