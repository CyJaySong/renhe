package rhttp

import (
	"github.com/cyjaysong/renhe/net/rhttp/middleware"
	"github.com/cyjaysong/renhe/os/rlog"
	"github.com/cyjaysong/renhe/os/rotrace"
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
	cfg := loadConfig()
	echoEngine := echo.New()
	echoEngine.HideBanner = true
	echoEngine.Validator = rvalid.Instance()
	echoEngine.Logger = rlog.Log().EchoLogger()
	echoEngine.Use(otelecho.Middleware(rotrace.ServiceName()))
	echoEngine.Use(middleware.ValidationMiddleware())
	return &HttpSrv{Echo: echoEngine, Cfg: cfg}
}
