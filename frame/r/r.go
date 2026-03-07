// Package r 是框架的门面包，提供全局组件的快捷访问入口。
// 通过 r.Cfg()、r.Log()、r.DB()、r.Redis()、r.HttpSrv() 获取各组件实例。
package r

import (
	"github.com/cyjaysong/renhe/database/rdb"
	"github.com/cyjaysong/renhe/database/redis"
	"github.com/cyjaysong/renhe/net/rhttp"
	"github.com/cyjaysong/renhe/os/rcfg"
	"github.com/cyjaysong/renhe/os/rlog"
	"github.com/spf13/viper"
)

// parseName 解析可选的实例名称参数，默认返回 "default"。
func parseName(name []string) string {
	if len(name) > 0 && name[0] != "" {
		return name[0]
	}
	return "default"
}

// HttpSrv 创建并返回 HTTP 服务器实例（内嵌 echo.Echo）。
func HttpSrv() *rhttp.HttpSrv {
	return rhttp.New()
}

// Cfg 返回全局配置实例，可传入路径片段获取子配置。
func Cfg(paths ...string) *viper.Viper {
	return rcfg.Cfg(paths...)
}

// Log 返回全局日志单例。
func Log() *rlog.Logger {
	return rlog.Log()
}

// DB 返回指定名称的数据库单例，默认 "default"。
func DB(name ...string) *rdb.DB {
	return rdb.Database(parseName(name))
}

// Redis 返回指定名称的 Redis 单例，默认 "default"。
func Redis(name ...string) *redis.Redis {
	return redis.Instance(parseName(name))
}
