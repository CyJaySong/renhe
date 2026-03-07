package rhttp

import (
	"log"

	"github.com/cyjaysong/renhe/os/rlog"
	"github.com/cyjaysong/renhe/util/rvalid"
	"github.com/labstack/echo/v4"
)

type HttpSrv struct {
	Cfg Config
	*echo.Echo
}

func New() *HttpSrv {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("http srv: %v\n", err)
		return nil
	}
	echoEngine := echo.New()
	echoEngine.Validator = rvalid.Instance()
	echoEngine.Logger = rlog.Log().EchoLogger()
	return &HttpSrv{Echo: echoEngine, Cfg: cfg}
}
