package cmd

import (
	"example/internal/controller"

	"github.com/cyjaysong/renhe/frame/r"
	"github.com/cyjaysong/renhe/net/rhttp"
	"github.com/cyjaysong/renhe/os/rctx"
	"github.com/labstack/echo/v5/middleware"
)

var Main = &struct {
	Run func()
}{
	Run: func() {
		// Trace 由配置 trace.enable 自动启停；资源统一 r.Close
		defer r.Close()

		httpSrv := r.HttpSrv()
		httpSrv.Use(
			middleware.Recover(),
			// 示例环境放宽 CORS；生产请改为具体源
			middleware.CORS("*"),
			// 将控制器返回的 *BizRes 原样 JSON 写出
			rhttp.WriteBizResJSON(),
		)

		api := httpSrv.Group("/api")
		rhttp.EchoRegisterCtrlPointers(api, new(controller.User))

		// 阻塞至 SIGINT/SIGTERM 优雅退出
		if err := httpSrv.Run(); err != nil {
			r.Log().Fatal(rctx.GetInitCtx(), "http server run failed", "err", err)
		}
	},
}
