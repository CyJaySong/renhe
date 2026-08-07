package rhttp_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cyjaysong/renhe/net/rhttp"
	"github.com/cyjaysong/renhe/util/rvalid"
	"github.com/labstack/echo/v5"
)

type helloReq struct {
	rhttp.HttpApiMeta `path:"/hello" method:"GET" name:"hello"`
}

type helloRes struct {
	Msg string `json:"msg"`
}

type helloCtrl struct{}

func (h *helloCtrl) Hello(ctx *echo.Context, req *helloReq) (*helloRes, error) {
	return &helloRes{Msg: "ok"}, nil
}

func TestEchoRegisterCtrlPointers_AndWriteBizRes(t *testing.T) {
	e := echo.NewWithConfig(echo.Config{Validator: rvalid.Instance()})
	e.Use(rhttp.WriteBizResJSON())
	api := e.Group("/api")
	rhttp.EchoRegisterCtrlPointers(api, new(helloCtrl))

	found := false
	for _, rt := range e.Router().Routes() {
		if rt.Path == "/api/hello" && rt.Name == "hello" && rt.Method == http.MethodGet {
			found = true
		}
	}
	if !found {
		t.Fatal("hello route not registered")
	}

	req := httptest.NewRequest(http.MethodGet, "/api/hello", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"msg":"ok"`) {
		t.Fatalf("body=%s", rec.Body.String())
	}
}
