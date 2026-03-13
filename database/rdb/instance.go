package rdb

import (
	"sync"

	"github.com/cyjaysong/renhe/os/rctx"

	"github.com/cyjaysong/renhe/os/rlog"
)

var (
	instances = make(map[string]*DB)
	mu        sync.RWMutex
)

// Database 返回指定名称的数据库单例（双检锁）。
// 不传参或传空字符串时使用 "default"。首次调用时自动加载配置并创建实例。
func Database(name ...string) *DB {
	n := "default"
	if len(name) > 0 && name[0] != "" {
		n = name[0]
	}

	mu.RLock()
	if d, ok := instances[n]; ok {
		mu.RUnlock()
		return d
	}
	mu.RUnlock()

	mu.Lock()
	defer mu.Unlock()
	if d, ok := instances[n]; ok {
		return d
	}
	cfg, err := loadConfig(n)
	if err != nil {
		rlog.Log().Error(rctx.GetInitCtx(), "rdb: load config failed", "name", n, "err", err)
		return nil
	}
	d, err := newDB(n, cfg)
	if err != nil {
		rlog.Log().Error(rctx.GetInitCtx(), "rdb: failed to create instance", "name", n, "err", err)
		return nil
	}
	instances[n] = d
	return d
}
