package custom

import (
	"unicode"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

// RegisterHanziValidation 向验证器注册 hanzi 标签及其中文翻译。
func RegisterHanziValidation(v *validator.Validate, zhTrans ut.Translator) (err error) {
	if err = v.RegisterValidation("hanzi", validateHanzi); err != nil {
		return
	}
	err = v.RegisterTranslation("hanzi", zhTrans,
		func(ut ut.Translator) error {
			return ut.Add("hanzi", "{0}必须全部为汉字", false)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("hanzi", fe.Field())
			return t
		},
	)
	return
}

// validateHanzi 校验字符串是否全部由汉字组成。
// 空字符串返回 false（需配合 required 使用或由 omitempty 跳过）。
func validateHanzi(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if !unicode.Is(unicode.Han, r) {
			return false
		}
	}
	return true
}
