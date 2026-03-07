package rhttp

import (
	"context"

	"github.com/cyjaysong/renhe/os/rlog"
	"github.com/cyjaysong/renhe/util/rvalid"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

// HttpSrv HTTP 服务器，内嵌 echo.Echo 并携带配置。
type HttpSrv struct {
	Cfg Config
	*echo.Echo
}

// New 创建 HttpSrv 实例，加载配置并初始化 echo 引擎、验证器和日志。
func New() *HttpSrv {
	cfg, err := loadConfig()
	if err != nil {
		rlog.Log().Fatal(context.Background(), "http srv: load config failed", "err", err)
		return nil
	}
	echoEngine := echo.New()
	echoEngine.Validator = rvalid.Instance()
	echoEngine.Logger = rlog.Log().EchoLogger()
	echoEngine.Use(otelecho.Middleware("Echo"))
	return &HttpSrv{Echo: echoEngine, Cfg: cfg}
}
