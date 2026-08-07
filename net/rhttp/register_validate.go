package rhttp

import (
	"fmt"
	"reflect"
	"strings"
)

// routeMethodMeta 校验通过后的路由元数据。
type routeMethodMeta struct {
	Name    string
	Path    string
	Methods []string // 空表示 Any
}

// validateCtrlMethod 严格校验 Ctrl 公开方法签名与 HttpApiMeta。
// 失败返回 error（供单测）；注册入口再 Fatal。
func validateCtrlMethod(funcType reflect.Type, funcFillName string) (routeMethodMeta, error) {
	var zero routeMethodMeta
	if funcType.NumIn() != 2 || funcType.NumOut() == 0 || funcType.NumOut() > 2 {
		return zero, fmt.Errorf("%s: check=1 %s", funcFillName, ctrlFuncSignHint)
	}
	if funcType.In(0) != echoContextPtrType {
		return zero, fmt.Errorf("%s: check=2 %s", funcFillName, ctrlFuncSignHint)
	}
	in1 := funcType.In(1)
	if in1.Kind() != reflect.Pointer || in1.Elem().Kind() != reflect.Struct {
		return zero, fmt.Errorf("%s: check=3 %s", funcFillName, ctrlFuncSignHint)
	}
	if funcType.NumOut() == 2 {
		if out := funcType.Out(0); out.Kind() != reflect.Pointer || out.Elem().Kind() != reflect.Struct {
			return zero, fmt.Errorf("%s: check=4 %s", funcFillName, ctrlFuncSignHint)
		}
	}
	if !funcType.Out(funcType.NumOut() - 1).Implements(errorReflectType) {
		return zero, fmt.Errorf("%s: check=5 %s", funcFillName, ctrlFuncSignHint)
	}

	var hasHttpApiMeta bool
	var nameTag, pathTag, methodTag string
	for field := range in1.Elem().Fields() {
		if field.Type == metaType {
			hasHttpApiMeta, nameTag = true, field.Tag.Get(metaNameTagName)
			pathTag, methodTag = field.Tag.Get(metaPathTagName), field.Tag.Get(metaMethodTagName)
			break
		}
	}
	if !hasHttpApiMeta {
		return zero, fmt.Errorf("%s: check=6 *BizReq 必须内嵌 rhttp.HttpApiMeta", funcFillName)
	}

	var methods []string
	if methodTag = strings.ToUpper(strings.TrimSpace(methodTag)); methodTag != "" {
		methods = strings.Split(methodTag, ",")
		for _, method := range methods {
			if _, has := allowMethod[method]; !has {
				return zero, fmt.Errorf("%s: method 标签不可用: %s", funcFillName, method)
			}
		}
	}
	return routeMethodMeta{Name: nameTag, Path: pathTag, Methods: methods}, nil
}
