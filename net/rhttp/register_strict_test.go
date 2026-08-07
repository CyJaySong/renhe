package rhttp_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cyjaysong/renhe/net/rhttp"
	"github.com/cyjaysong/renhe/util/rvalid"
	"github.com/labstack/echo/v5"
)

type onlyRouteReq struct {
	rhttp.HttpApiMeta `path:"/only" method:"GET" name:"only"`
}

type onlyRouteRes struct {
	OK bool `json:"ok"`
}

type onlyRouteCtrl struct{}

func (c *onlyRouteCtrl) Only(ctx *echo.Context, req *onlyRouteReq) (*onlyRouteRes, error) {
	return &onlyRouteRes{OK: true}, nil
}

// 不存在辅助公开方法：所有公开方法均可注册
func TestStrictCtrl_AllPublicMethodsAreRoutes(t *testing.T) {
	e := echo.NewWithConfig(echo.Config{Validator: rvalid.Instance()})
	e.Use(rhttp.WriteBizResJSON())
	api := e.Group("/api")
	rhttp.EchoRegisterCtrlPointers(api, new(onlyRouteCtrl))

	n := 0
	for _, rt := range e.Router().Routes() {
		if rt.Path == "/api/only" && rt.Method == http.MethodGet {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("routes=%d", n)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/only", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
}
