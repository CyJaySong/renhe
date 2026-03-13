package rvalid

// Validate 使用 "v" tag 校验结构体。实现 echo.Validator 接口。
func (v *Validator) Validate(obj any) error {
	return v.v1.Struct(obj)
}
