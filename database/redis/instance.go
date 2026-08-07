package redis

import (
	"sync"

	"github.com/cyjaysong/renhe/os/rctx"
	"github.com/cyjaysong/renhe/os/rlog"
)

var (
	instances = make(map[string]*Redis)
	mu        sync.RWMutex
)

// Database 返回指定名称的 Redis 单例（双检锁）。
// 不传参或传空字符串时使用 "default"。首次调用时自动加载配置并创建实例。
// 配置或连接失败时记 Error 日志并返回 nil。
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
	cfg, err := loadConfig(n)
	if err != nil {
		rlog.Log().Error(rctx.GetInitCtx(), "redis: load config failed", "name", n, "err", err)
		return nil
	}
	r, err := newRedis(n, cfg)
	if err != nil {
		rlog.Log().Error(rctx.GetInitCtx(), "redis: failed to create instance", "name", n, "err", err)
		return nil
	}
	instances[n] = r
	return r
}

// CloseAll 关闭所有已创建的 Redis 实例并清空单例缓存。
func CloseAll() {
	mu.Lock()
	defer mu.Unlock()
	for name, r := range instances {
		if r != nil && r.UniversalClient != nil {
			if err := r.Close(); err != nil {
				rlog.Log().Error(rctx.GetInitCtx(), "redis: close failed", "name", name, "err", err)
			}
		}
		delete(instances, name)
	}
}
