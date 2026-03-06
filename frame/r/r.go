package r

import (
	"github.com/cyjaysong/renhe/database/rdb"
	"github.com/cyjaysong/renhe/os/rcfg"
	"github.com/cyjaysong/renhe/os/rlog"
	"github.com/cyjaysong/renhe/util/rvalid"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

func parseName(name []string) string {
	if len(name) > 0 && name[0] != "" {
		return name[0]
	}
	return "default"
}

func HttpSrv() *echo.Echo {
	httpEngine := echo.New()
	httpEngine.Validator = rvalid.Instance()
	return httpEngine
}

func Cfg(name ...string) *viper.Viper {
	return rcfg.Instance(parseName(name))
}

func Log(name ...string) *rlog.Logger {
	return rlog.Instance(parseName(name))
}

func DB(name ...string) *rdb.DB {
	return rdb.Instance(parseName(name))
}
