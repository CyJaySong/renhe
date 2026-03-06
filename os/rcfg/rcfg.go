package rcfg

import (
	"sync"

	"github.com/spf13/viper"
)

var (
	instances = make(map[string]*viper.Viper)
	mu        sync.RWMutex
)

func Instance(name ...string) *viper.Viper {
	n := "default"
	if len(name) > 0 && name[0] != "" {
		n = name[0]
	}
	mu.RLock()
	if v, ok := instances[n]; ok {
		mu.RUnlock()
		return v
	}
	mu.RUnlock()

	mu.Lock()
	defer mu.Unlock()
	if v, ok := instances[n]; ok {
		return v
	}
	v := viper.New()
	v.SetConfigType("yaml")
	if n == "default" {
		v.SetConfigName("config")
	} else {
		v.SetConfigName(n)
	}
	v.AddConfigPath("manifest/config")
	v.AddConfigPath("config")
	v.AddConfigPath(".")
	_ = v.ReadInConfig()
	instances[n] = v
	return v
}
