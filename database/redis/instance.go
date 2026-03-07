package redis

import (
	"fmt"
	"sync"
)

var (
	instances = make(map[string]*Redis)
	mu        sync.RWMutex
)

func Instance(name ...string) *Redis {
	n := "default"
	if len(name) > 0 && name[0] != "" {
		n = name[0]
	}

	mu.RLock()
	if r, ok := instances[n]; ok {
		mu.RUnlock()
		return r
	}
	mu.RUnlock()

	mu.Lock()
	defer mu.Unlock()
	if r, ok := instances[n]; ok {
		return r
	}
	cfg, err := loadConfig(n)
	if err != nil {
		fmt.Printf("redis: %v\n", err)
		return nil
	}
	r, err := newRedis(n, cfg)
	if err != nil {
		fmt.Printf("redis: failed to create instance %q: %v\n", n, err)
		return nil
	}
	instances[n] = r
	return r
}
