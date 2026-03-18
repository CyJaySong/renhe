package custom

import (
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
)

// RegisterDecimalType 向验证器注册 decimal.Decimal 的自定义类型转换，
// 使内置的 required、gt、gte、lt、lte、eq、ne 等数值类 tag 直接生效。
func RegisterDecimalType(v *validator.Validate) {
	v.RegisterCustomTypeFunc(decimalTypeFunc, decimal.Decimal{})
}

// decimalTypeFunc 将 decimal.Decimal 转换为 float64，供 validator 内置比较逻辑使用。
func decimalTypeFunc(field reflect.Value) any {
	if d, ok := field.Interface().(decimal.Decimal); ok {
		f, _ := d.Float64()
		return f
	}
	return nil
}
