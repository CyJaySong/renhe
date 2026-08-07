package custom

import (
	"reflect"

	"github.com/go-playground/validator/v10"
	"github.com/shopspring/decimal"
)

// RegisterDecimalType 向验证器注册 decimal.Decimal, decimal.NullDecimal 的自定义类型转换，
// 使内置的 required、gt、gte、lt、lte、eq、ne 等数值类 tag 直接生效。
func RegisterDecimalType(v *validator.Validate) {
	v.RegisterCustomTypeFunc(decimalTypeFunc, decimal.Decimal{})
	v.RegisterCustomTypeFunc(decimalTypeFunc, decimal.NullDecimal{})
}

// decimalTypeFunc 将 decimal.Decimal 转换为 float64，供 validator 内置比较逻辑使用。
func decimalTypeFunc(field reflect.Value) any {
	switch v := field.Interface().(type) {
	case decimal.Decimal:
		f, _ := v.Float64()
		return f
	case decimal.NullDecimal:
		if v.Valid {
			f, _ := v.Decimal.Float64()
			return f
		}
	}
	return nil
}
