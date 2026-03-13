// Package rvalid 提供参数验证服务，基于 go-playground/validator 实现。
// 内置两组验证器（v tag 和 v2 tag），通过 Instance() 获取单例。
package rvalid

import (
	"github.com/go-playground/validator/v10"
	zhTrans "github.com/go-playground/validator/v10/translations/zh"
	"sync"
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
		_ = zhTrans.RegisterDefaultTranslations(v1, zhTranslator)
		v1.SetTagName("v")

		v2 := validator.New(validator.WithRequiredStructEnabled())
		_ = zhTrans.RegisterDefaultTranslations(v2, zhTranslator)
		v2.SetTagName("v2")

		instance = &Validator{v1: v1, v2: v2}
	})
	return instance
}

func (v *Validator) V() *validator.Validate {
	return v.v1
}

func (v *Validator) V2() *validator.Validate {
	return v.v2
}
