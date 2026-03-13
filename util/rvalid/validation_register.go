package rvalid

import "github.com/go-playground/validator/v10"

// RegisterValidationCtx 同时向两组验证器注册自定义校验规则。
func (v *Validator) RegisterValidationCtx(tag string, funcCtx validator.FuncCtx) (err error) {
	if err = v.v1.RegisterValidationCtx(tag, funcCtx); err != nil {
		return
	}
	return v.v2.RegisterValidationCtx(tag, funcCtx)
}
