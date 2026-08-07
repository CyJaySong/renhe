package rhttp

import (
	"reflect"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
)

type goodReq struct {
	HttpApiMeta `path:"/g" method:"GET" name:"g"`
}
type goodRes struct{}

func goodHandler(c *echo.Context, req *goodReq) (*goodRes, error) { return &goodRes{}, nil }

func badHelper() string { return "x" }

type noMetaReq struct {
	X int
}

func noMetaHandler(c *echo.Context, req *noMetaReq) error { return nil }

func TestValidateCtrlMethod_OK(t *testing.T) {
	ft := reflect.TypeOf(goodHandler)
	meta, err := validateCtrlMethod(ft, "good")
	if err != nil {
		t.Fatal(err)
	}
	if meta.Path != "/g" || meta.Name != "g" || len(meta.Methods) != 1 || meta.Methods[0] != "GET" {
		t.Fatalf("%+v", meta)
	}
}

func TestValidateCtrlMethod_RejectHelper(t *testing.T) {
	ft := reflect.TypeOf(badHelper)
	_, err := validateCtrlMethod(ft, "Helper")
	if err == nil || !strings.Contains(err.Error(), "check=1") {
		t.Fatalf("want check=1, got %v", err)
	}
}

func TestValidateCtrlMethod_RejectNoMeta(t *testing.T) {
	ft := reflect.TypeOf(noMetaHandler)
	_, err := validateCtrlMethod(ft, "NoMeta")
	if err == nil || !strings.Contains(err.Error(), "check=6") {
		t.Fatalf("want check=6, got %v", err)
	}
}
