package redis

import (
	"github.com/cyjaysong/renhe/os/rctx"
	"sync"

	"github.com/cyjaysong/renhe/os/rlog"
)

var (
	instances = make(map[string]*Redis)
	mu        sync.RWMutex
)

// Database 返回指定名称的 Redis 单例（双检锁）。
// 不传参或传空字符串时使用 "default"。首次调用时自动加载配置并创建实例。
func Database(name ...string) *Redis {
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
	cfg := loadConfig(n)
	r, err := newRedis(n, cfg)
	if err != nil {
		rlog.Log().Errorf(rctx.GetInitCtx(), "redis: failed to create `%s` instance: %s", name, err)
		return nil
	}
	instances[n] = r
	return r
}
