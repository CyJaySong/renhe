package rhttp

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/cyjaysong/renhe/os/rctx"
	"github.com/cyjaysong/renhe/os/rlog"
	"github.com/cyjaysong/renhe/util/rvalid"
	"github.com/labstack/echo/v5"
)

const echoCtrlFuncBizResKey = "echo_ctrl_func_biz_res"

var errorReflectType = reflect.TypeOf((*error)(nil)).Elem()

// echoContextPtrType 为 *echo.Context（v5 Context 是结构体，不再是 interface）
var echoContextPtrType = reflect.TypeOf((*echo.Context)(nil))

var allowMethod = map[string]struct{}{
	http.MethodGet:     {},
	http.MethodHead:    {},
	http.MethodPost:    {},
	http.MethodPut:     {},
	http.MethodPatch:   {},
	http.MethodDelete:  {},
	http.MethodConnect: {},
	http.MethodOptions: {},
	http.MethodTrace:   {},
}

// 签名不符时的统一提示
const ctrlFuncSignHint = "函数签名不符, 应为 func(*echo.Context, *BizReq)(error) 或 func(*echo.Context, *BizReq)(*BizRes, error)"

// EchoRegisterCtrlPointers 接受一个 echo 引擎或者 echo 路由组，和多个 ctrl 指针对象。
// Ctrl 上每一个公开方法都必须是合法路由签名，且 *BizReq 内嵌 HttpApiMeta；不允许辅助公开方法。
func EchoRegisterCtrlPointers[T *echo.Echo | *echo.Group](echoOrGroup T, ctrlPointers ...any) {
	for _, ctrl := range ctrlPointers {
		ctrlType := reflect.TypeOf(ctrl)
		if ctrlType.Kind() != reflect.Pointer || ctrlType.Elem().Kind() != reflect.Struct {
			rlog.Log().Fatal(rctx.GetInitCtx(), "用于注册的ctrl对象必须*Ctrl形式", "ctrl", ctrlType.Name())
		}
		ctrlTypePath, ctrlTypeName := ctrlType.Elem().PkgPath(), ctrlType.Elem().Name()
		ctrlValue := reflect.ValueOf(ctrl)
		for i := range ctrlValue.NumMethod() {
			funcFillName := fmt.Sprintf("%s.%s.%s", ctrlTypePath, ctrlTypeName, ctrlType.Method(i).Name)
			funcItem := ctrlValue.Method(i)
			funcType := funcItem.Type()

			meta, err := validateCtrlMethod(funcType, funcFillName)
			if err != nil {
				rlog.Log().Fatal(rctx.GetInitCtx(), err.Error())
			}
			echoHandlerRegister(echoOrGroup, meta.Methods, meta.Path, meta.Name, funcType, funcItem)
		}
	}
}

// GetCtrlFuncBizRes 从 *echo.Context 获取使用 Ctrl 注册的 BizRes
func GetCtrlFuncBizRes(ctx *echo.Context) any {
	return ctx.Get(echoCtrlFuncBizResKey)
}

// routeAdder Echo / Group 统一注册入口
type routeAdder interface {
	AddRoute(route echo.Route) (echo.RouteInfo, error)
}

func echoHandlerRegister[T *echo.Echo | *echo.Group](echoOrGroup T, methods []string, path, name string,
	funcType reflect.Type, funcItem reflect.Value) {
	adder := any(echoOrGroup).(routeAdder)
	h := echoHandler(funcType, funcItem)
	if len(methods) == 0 {
		methods = []string{echo.RouteAny}
	}
	for _, method := range methods {
		if _, err := adder.AddRoute(echo.Route{Method: method, Path: path, Name: name, Handler: h}); err != nil {
			rlog.Log().Fatal(rctx.GetInitCtx(), "路由注册失败", "method", method, "path", path, "name", name, "err", err)
		}
	}
}

func echoHandler(funcType reflect.Type, funcItem reflect.Value) echo.HandlerFunc {
	return func(ctx *echo.Context) (err error) {
		bizReq := reflect.New(funcType.In(1).Elem())
		if err = echo.BindHeaders(ctx, bizReq.Interface()); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error()).Wrap(err)
		}
		if err = ctx.Bind(bizReq.Interface()); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error()).Wrap(err)
		}
		if err = ctx.Validate(bizReq.Interface()); err != nil {
			msg := rvalid.FirstError(err)
			if msg == "" {
				msg = err.Error()
			}
			return echo.NewHTTPError(http.StatusBadRequest, msg).Wrap(err)
		}
		inParams := []reflect.Value{reflect.ValueOf(ctx), bizReq}
		outValues := funcItem.Call(inParams)

		if len(outValues) == 2 && outValues[0].IsValid() && !outValues[0].IsNil() {
			ctx.Set(echoCtrlFuncBizResKey, outValues[0].Interface())
		}
		if errValue := outValues[len(outValues)-1]; !errValue.IsNil() {
			err = errValue.Interface().(error)
		}
		return
	}
}
