package redis

import (
	"fmt"
	"sync"
)

var (
	instances = make(map[string]*Redis)
	mu        sync.RWMutex
)

// Instance 返回指定名称的 Redis 单例（双检锁）。
// 不传参或传空字符串时使用 "default"。首次调用时自动加载配置并创建实例。
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
