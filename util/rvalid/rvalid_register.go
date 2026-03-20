package rvalid

import (
	"github.com/go-playground/validator/v10"
)

// RegisterCustomTypeFunc 同时向两组验证器注册自定义类型转换函数。
func (v *Validator) RegisterCustomTypeFunc(fn validator.CustomTypeFunc, types ...any) {
	v.v1.RegisterCustomTypeFunc(fn, types...)
	v.v2.RegisterCustomTypeFunc(fn, types...)
}

// RegisterValidationCtx 同时向两组验证器注册自定义校验规则。
func (v *Validator) RegisterValidationCtx(tag string, funcCtx validator.FuncCtx) (err error) {
	if err = v.v1.RegisterValidationCtx(tag, funcCtx); err != nil {
		return
	}
	return v.v2.RegisterValidationCtx(tag, funcCtx)
}

// RegisterStructValidation 同时向两组验证器注册结构体级别的校验函数。
func (v *Validator) RegisterStructValidation(fn validator.StructLevelFunc, types ...any) {
	v.v1.RegisterStructValidation(fn, types...)
	v.v2.RegisterStructValidation(fn, types...)
}

// RegisterZhTranslation 同时向两组验证器注册中文翻译。
func (v *Validator) RegisterZhTranslation(tag string, registerFn validator.RegisterTranslationsFunc, translationFn validator.TranslationFunc) (err error) {
	if err = v.v1.RegisterTranslation(tag, zhTranslator1, registerFn, translationFn); err != nil {
		return
	}
	return v.v2.RegisterTranslation(tag, zhTranslator2, registerFn, translationFn)
}
