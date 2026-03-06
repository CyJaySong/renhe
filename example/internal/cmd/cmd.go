package cmd

import (
	"example/internal/controller"
	"github.com/cyjaysong/renhe/frame/r"
	"github.com/cyjaysong/renhe/net/rhttp"
	"github.com/labstack/echo/v4/middleware"
)

var Main = &struct {
	Run func()
}{
	Run: func() {
		httpSrv := r.HttpSrv()
		httpSrv.Use(middleware.Recover())

		api := httpSrv.Group("/api")
		rhttp.EchoRegisterCtrlPointers(api, new(controller.User))

		httpSrv.Logger.Fatal(httpSrv.Start(":8000"))
	},
}
