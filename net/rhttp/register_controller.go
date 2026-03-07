package rhttp

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/cyjaysong/renhe/os/rlog"
	"github.com/labstack/echo/v4"
)

const echoCtrlFuncBizResKey = "echo_ctrl_func_biz_res"

var errorReflectType = reflect.TypeOf((*error)(nil)).Elem()
var echoContextReflectType = reflect.TypeOf((*echo.Context)(nil)).Elem()

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

// EchoRegisterCtrlPointers 接受一个echo引擎或者echo路由组，和多个ctrl指针对象，判断各ctrl指针对象的所有函数，是否可用于注册
func EchoRegisterCtrlPointers[T *echo.Echo | *echo.Group](echoOrGroup T, ctrlPointers ...any) {
	for _, ctrl := range ctrlPointers {
		ctrlType := reflect.TypeOf(ctrl)
		if ctrlType.Kind() != reflect.Pointer || ctrlType.Elem().Kind() != reflect.Struct {
			rlog.Log().Fatal(context.Background(), "用于注册的ctrl对象必须*Ctrl形式", "ctrl", ctrlType.Name())
		}
		ctrlTypePath, ctrlTypeName := ctrlType.Elem().PkgPath(), ctrlType.Elem().Name()
		ctrlValue := reflect.ValueOf(ctrl)
		for i := 0; i < ctrlValue.NumMethod(); i++ {
			// 标准的 reflect 包无法直接获取非导出方法, 固不用担心遍历到非导出方法
			funcFillName := fmt.Sprintf("%s.%s.%s", ctrlTypePath, ctrlTypeName, ctrlType.Method(i).Name)
			funcItem := ctrlValue.Method(i)
			funcType := funcItem.Type()
			if funcType.NumIn() != 2 || funcType.NumOut() == 0 || funcType.NumOut() > 2 {
				rlog.Log().Fatal(context.Background(), "函数签名不符, 应为 func(echo.Context, *BizReq)(error) 或 func(echo.Context, *BizReq)(*BizRes, error)", "func", funcFillName, "check", 1)
			}
			// 入参判断
			if !funcType.In(0).Implements(echoContextReflectType) {
				rlog.Log().Fatal(context.Background(), "函数签名不符, 应为 func(echo.Context, *BizReq)(error) 或 func(echo.Context, *BizReq)(*BizRes, error)", "func", funcFillName, "check", 2)
			} else if in := funcType.In(1); in.Kind() != reflect.Pointer || in.Elem().Kind() != reflect.Struct {
				rlog.Log().Fatal(context.Background(), "函数签名不符, 应为 func(echo.Context, *BizReq)(error) 或 func(echo.Context, *BizReq)(*BizRes, error)", "func", funcFillName, "check", 3)
			}
			// 出参有2个时参数判断第一个
			if funcType.NumOut() == 2 {
				if out := funcType.Out(0); out.Kind() != reflect.Pointer || out.Elem().Kind() != reflect.Struct {
					rlog.Log().Fatal(context.Background(), "函数签名不符, 应为 func(echo.Context, *BizReq)(error) 或 func(echo.Context, *BizReq)(*BizRes, error)", "func", funcFillName, "check", 4)
				}
			}
			// 出参error判断 funcType.NumOut() == 1 or funcType.NumOut() == 2
			if !funcType.Out(funcType.NumOut() - 1).Implements(errorReflectType) {
				rlog.Log().Fatal(context.Background(), "函数签名不符, 应为 func(echo.Context, *BizReq)(error) 或 func(echo.Context, *BizReq)(*BizRes, error)", "func", funcFillName, "check", 5)
			}

			// 判断第二个入参是否包含 HttpApiMeta 元数据, 没有则跳过
			var hasHttpApiMeta bool
			var nameTag, pathTag, methodTag string
			for field := range funcType.In(1).Elem().Fields() {
				if field.Type == metaType {
					hasHttpApiMeta, nameTag = true, field.Tag.Get(metaNameTagName)
					pathTag, methodTag = field.Tag.Get(metaPathTagName), field.Tag.Get(metaMethodTagName)
					break
				}
			}
			if !hasHttpApiMeta {
				continue
			}

			var methods []string
			if methodTag = strings.ToUpper(strings.TrimSpace(methodTag)); methodTag != "" {
				methods = strings.Split(methodTag, ",")
				disableMethods := make([]string, 0, len(methods))
				for _, method := range methods {
					if _, has := allowMethod[method]; !has {
						disableMethods = append(disableMethods, method)
					}
				}
				if len(disableMethods) > 0 {
					rlog.Log().Fatal(context.Background(), "函数签名不符, *BizReq method标签值不可用", "func", funcFillName, "disableMethods", strings.Join(disableMethods, ","))
				}
			}
			echoHandlerRegister(echoOrGroup, methods, pathTag, nameTag, funcType, funcItem)
		}
	}
}

// GetCtrlFuncBizRes 从echo.Context获取使用Ctrl注册的BizRes
func GetCtrlFuncBizRes(ctx echo.Context) any {
	return ctx.Get(echoCtrlFuncBizResKey)
}

type echoEngineOrGroup interface {
	Any(path string, handler echo.HandlerFunc, middleware ...echo.MiddlewareFunc) []*echo.Route
	Match(methods []string, path string, handler echo.HandlerFunc, middleware ...echo.MiddlewareFunc) []*echo.Route
}

func echoHandlerRegister[T *echo.Echo | *echo.Group](echoOrGroup T, methods []string, path, name string,
	funcType reflect.Type, funcItem reflect.Value) {
	engine := any(echoOrGroup).(echoEngineOrGroup)
	if len(methods) == 0 {
		for _, route := range engine.Any(path, echoHandler(funcType, funcItem)) {
			route.Name = name
		}
	} else {
		for _, route := range engine.Match(methods, path, echoHandler(funcType, funcItem)) {
			route.Name = name
		}
	}
}

func echoHandler(funcType reflect.Type, funcItem reflect.Value) echo.HandlerFunc {
	return func(ctx echo.Context) (err error) {
		bizReq := reflect.New(funcType.In(1).Elem())
		if err = ctx.Bind(bizReq.Interface()); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		// 校验入参
		if err = ctx.Validate(bizReq.Interface()); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		// 构建入参
		inParams := []reflect.Value{reflect.ValueOf(ctx), bizReq}
		// 调用函数
		outValues := funcItem.Call(inParams)

		if len(outValues) == 2 && outValues[0].IsValid() && !outValues[0].IsNil() {
			ctx.Set(echoCtrlFuncBizResKey, outValues[0].Interface())
		}
		// 获取处理函数返回的error
		if errValue := outValues[len(outValues)-1]; !errValue.IsNil() {
			err = errValue.Interface().(error)
		}
		//echo.NewHTTPError()
		//if len(outValues) == 2 && outValues[1].IsNil() {
		//	var data any
		//	if outValues[0].IsValid() && !outValues[0].IsNil() {
		//		data = outValues[0].Interface()
		//	} else {
		//		data = struct{}{}
		//	}
		//	//echo.HTTPError{}
		//	return ctx.JSON(http.StatusOK, &api.BaseApiResponseBody{
		//		Code:    0,
		//		Message: "",
		//		Data:    data,
		//	})
		//}
		//// 处理返回值
		//if !outValues[1].IsNil() {
		//	return outValues[1].Interface().(error)
		//}
		//if outValues[0].IsValid() {
		//	return ctx.JSON(200, outValues[0].Interface())
		//}
		return
	}
}
