package rhttp

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/cyjaysong/renhe/os/rctx"
	"github.com/cyjaysong/renhe/os/rlog"
	"github.com/cyjaysong/renhe/os/rotrace"
	"github.com/cyjaysong/renhe/util/rvalid"
	echootel "github.com/labstack/echo-opentelemetry"
	"github.com/labstack/echo/v5"
)

// HttpSrv HTTP 服务器，内嵌 echo.Echo 并携带配置。
type HttpSrv struct {
	Cfg Config
	*echo.Echo
}

// New 创建 HttpSrv 实例，加载配置并初始化 echo 引擎、验证器和日志。
// 配置缺失或 address 为空时 Fatal。
// 若 trace.enable=true：自动 rotrace.Ensure，并挂 OTel HTTP 中间件。
func New() *HttpSrv {
	cfg, err := loadConfig()
	if err != nil {
		rlog.Log().Fatal(rctx.GetInitCtx(), "rhttp: load config failed", "err", err)
	}
	// 按配置自动初始化 Trace（无需下游手写 InitTrace）
	rotrace.Ensure(rctx.GetInitCtx())
	// 使用 NewWithConfig 注入 slog Logger 与校验器
	echoEngine := echo.NewWithConfig(echo.Config{
		Logger:    rlog.Log().Slog(),
		Validator: rvalid.Instance(),
	})
	// OTel HTTP 中间件仅跟随 trace.enable
	if rotrace.Enabled() {
		echoEngine.Use(echootel.NewMiddleware(rotrace.ServiceName()))
	}
	return &HttpSrv{Echo: echoEngine, Cfg: cfg}
}

// Run 启动 HTTP 服务，监听 SIGINT/SIGTERM 并优雅退出。
// 正常关闭（含 http.ErrServerClosed）返回 nil；其它错误原样返回。
// 调用方应在 Run 返回后执行 r.Close() 等资源清理（可用 defer）。
func (s *HttpSrv) Run() error {
	return s.RunContext(context.Background())
}

// RunContext 与 Run 相同，但使用调用方提供的 parent context（可叠加超时/取消）。
func (s *HttpSrv) RunContext(parent context.Context) error {
	if parent == nil {
		parent = context.Background()
	}
	// 信号触发取消 → StartConfig 进入优雅关闭
	ctx, stop := signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM)
	defer stop()

	sc := echo.StartConfig{
		Address:         s.Cfg.Address,
		HideBanner:      true,
		GracefulTimeout: s.Cfg.gracefulTimeout(),
		OnShutdownError: func(err error) {
			rlog.Log().Error(rctx.GetInitCtx(), "rhttp: graceful shutdown error", "err", err)
		},
	}
	err := sc.Start(ctx, s.Echo)
	return normalizeStartErr(err)
}

// normalizeStartErr 将正常关闭视为成功。
func normalizeStartErr(err error) error {
	if err == nil || errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
