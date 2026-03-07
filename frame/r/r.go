package r

import (
	"github.com/cyjaysong/renhe/database/rdb"
	"github.com/cyjaysong/renhe/database/redis"
	"github.com/cyjaysong/renhe/net/rhttp"
	"github.com/cyjaysong/renhe/os/rcfg"
	"github.com/cyjaysong/renhe/os/rlog"
	"github.com/spf13/viper"
)

func parseName(name []string) string {
	if len(name) > 0 && name[0] != "" {
		return name[0]
	}
	return "default"
}

func HttpSrv() *rhttp.HttpSrv {
	return rhttp.New()
}

func Cfg(paths ...string) *viper.Viper {
	return rcfg.Cfg(paths...)
}

func Log() *rlog.Logger {
	return rlog.Log()
}

func DB(name ...string) *rdb.DB {
	return rdb.Database(parseName(name))
}

func Redis(name ...string) *redis.Redis {
	return redis.Instance(parseName(name))
}
