package custom

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

// RegisterDigitsValidation 向验证器注册 digits 标签及其中文翻译。
func RegisterDigitsValidation(v *validator.Validate, zhTrans ut.Translator) (err error) {
	if err = v.RegisterValidation("digits", validateDigits); err != nil {
		return
	}
	err = v.RegisterTranslation("digits", zhTrans,
		func(ut ut.Translator) error {
			return ut.Add("digits", "{0}的位数限制为{1}", false)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			param := fe.Param()
			maxInt, maxFrac := parseDigitsParam(param)
			var desc string
			switch {
			case maxInt == 0 && maxFrac == 0:
				desc = "仅允许值为0"
			case maxInt == 0 && maxFrac > 0:
				desc = fmt.Sprintf("整数部分必须为0，小数部分最多%d位", maxFrac)
			case maxInt > 0 && maxFrac == 0:
				desc = fmt.Sprintf("整数部分最多%d位，不允许小数部分", maxInt)
			case maxInt >= 0 && maxFrac >= 0:
				desc = fmt.Sprintf("整数部分最多%d位，小数部分最多%d位", maxInt, maxFrac)
			case maxInt == 0:
				desc = "整数部分必须为0"
			case maxInt > 0:
				desc = fmt.Sprintf("整数部分最多%d位", maxInt)
			case maxFrac == 0:
				desc = "不允许小数部分"
			case maxFrac > 0:
				desc = fmt.Sprintf("小数部分最多%d位", maxFrac)
			default:
				desc = param
			}
			t, _ := ut.T("digits", fe.Field(), desc)
			return t
		},
	)
	return
}

// validateDigits 校验数值的整数位和小数位长度。
//
// 参数格式: digits=整数位.小数位
//   - digits=10.2  整数部分最多10位，小数部分最多2位
//   - digits=.2    不限整数位，小数部分最多2位
//   - digits=10.   整数部分最多10位，不限小数位
//   - digits=10    整数部分最多10位，不限小数位（同上）
//   - digits=0.2   整数部分为0（仅允许 0.xx 形式）
//   - digits=2.0   小数部分为0（仅允许整数形式）
//
// 支持类型: float32、float64、int 系列（自定义类型经 CustomTypeFunc 转为底层数值类型后同样适用）
func validateDigits(fl validator.FieldLevel) bool {
	param := fl.Param()
	maxInt, maxFrac := parseDigitsParam(param)

	field := fl.Field()
	var f float64
	switch field.Kind() {
	case reflect.Float32, reflect.Float64:
		f = field.Float()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		f = float64(field.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		f = float64(field.Uint())
	default:
		return false
	}

	intLen, fracLen := floatDigits(f)

	if maxInt >= 0 && intLen > maxInt {
		return false
	}
	if maxFrac >= 0 && fracLen > maxFrac {
		return false
	}
	return true
}

// parseDigitsParam 解析 digits 参数。返回 (maxInt, maxFrac)，-1 表示不限制。
func parseDigitsParam(param string) (maxInt, maxFrac int) {
	maxInt, maxFrac = -1, -1
	idx := strings.IndexByte(param, '.')
	if idx < 0 {
		// digits=10 → 仅限整数位
		if n, err := strconv.Atoi(param); err == nil {
			maxInt = n
		}
		return
	}
	// 点号左侧
	if idx > 0 {
		if n, err := strconv.Atoi(param[:idx]); err == nil {
			maxInt = n
		}
	}
	// 点号右侧
	if idx < len(param)-1 {
		if n, err := strconv.Atoi(param[idx+1:]); err == nil {
			maxFrac = n
		}
	}
	return
}

// floatDigits 计算 float64 的整数位数和有效小数位数（去除尾零）。
// 负数取绝对值计算。"0" 的整数位数记为 0。
func floatDigits(f float64) (intLen, fracLen int) {
	f = math.Abs(f)

	// 用 strconv.FormatFloat 得到去除尾零的十进制表示
	s := strconv.FormatFloat(f, 'f', -1, 64)

	dotIdx := strings.IndexByte(s, '.')
	if dotIdx < 0 {
		intLen = countIntDigits(s)
		return
	}
	fracLen = len(s) - dotIdx - 1
	intLen = countIntDigits(s[:dotIdx])
	return
}

// countIntDigits 计算整数字符串的有效位数。"0" 返回 0。
func countIntDigits(s string) int {
	if s == "0" || s == "" {
		return 0
	}
	return len(s)
}
