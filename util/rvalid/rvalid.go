// Package rvalid 提供参数验证服务，基于 go-playground/validator 实现。
// 内置两组验证器（v tag 和 v2 tag），通过 Instance() 获取单例。
package rvalid

import (
	"sync"

	"github.com/go-playground/validator/v10"
)

var (
	instance *Validator
	once     sync.Once
)

// Validator 封装两组 go-playground/validator 实例，分别使用 "v" 和 "v2" tag。
type Validator struct {
	v1 *validator.Validate
	v2 *validator.Validate
}

// Instance 返回全局 Validator 单例。
func Instance() *Validator {
	once.Do(func() {
		v1 := validator.New(validator.WithRequiredStructEnabled())
		v1.SetTagName("v")

		v2 := validator.New(validator.WithRequiredStructEnabled())
		v2.SetTagName("v2")

		instance = &Validator{v1: v1, v2: v2}
	})
	return instance
}

// RegisterValidationCtx 同时向两组验证器注册自定义校验规则。
func (v *Validator) RegisterValidationCtx(tag string, funcCtx validator.FuncCtx) (err error) {
	if err = v.v1.RegisterValidationCtx(tag, funcCtx); err != nil {
		return
	}
	return v.v2.RegisterValidationCtx(tag, funcCtx)
}

// Validate 使用 "v" tag 校验结构体。实现 echo.Validator 接口。
func (v *Validator) Validate(obj any) error {
	return v.v1.Struct(obj)
}

// ValidateV2 使用 "v2" tag 校验结构体。
func (v *Validator) ValidateV2(obj any) error {
	return v.v2.Struct(obj)
}

// Validate 包级快捷方法，使用 "v" tag 校验。
func Validate(obj any) error {
	return Instance().Validate(obj)
}

// ValidateV2 包级快捷方法，使用 "v2" tag 校验。
func ValidateV2(obj any) error {
	return Instance().ValidateV2(obj)
}
