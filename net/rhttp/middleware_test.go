package rhttp_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cyjaysong/renhe/net/rhttp"
	"github.com/labstack/echo/v5"
)

func TestWriteBizResJSON_OnError_NoWrite(t *testing.T) {
	e := echo.New()
	e.Use(rhttp.WriteBizResJSON())
	e.GET("/err", func(c *echo.Context) error {
		// 即使 Set 了 BizRes，有 error 也不应写出
		c.Set("echo_ctrl_func_biz_res", map[string]string{"x": "y"})
		return errors.New("boom")
	})

	req := httptest.NewRequest(http.MethodGet, "/err", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	// 默认错误处理，body 不应是我们的 biz res
	if rec.Body.String() == `{"x":"y"}` {
		t.Fatalf("should not write biz res on error: %s", rec.Body.String())
	}
}

func TestWriteBizResJSON_NoBizRes_EmptyOK(t *testing.T) {
	e := echo.New()
	e.Use(rhttp.WriteBizResJSON())
	e.GET("/ok", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status=%d", rec.Code)
	}
}
