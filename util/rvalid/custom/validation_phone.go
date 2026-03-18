package custom

import (
	"regexp"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

var cnPhoneRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)

// RegisterPhoneValidation 向验证器注册 cnphone 标签及其中文翻译。
func RegisterPhoneValidation(v *validator.Validate, zhTrans ut.Translator) (err error) {
	if err = v.RegisterValidation("cnphone", validateCnPhone); err != nil {
		return
	}
	err = v.RegisterTranslation("cnphone", zhTrans,
		func(ut ut.Translator) error {
			return ut.Add("cnphone", "{0}必须为有效的中国手机号", false)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("cnphone", fe.Field())
			return t
		},
	)
	return
}

// validateCnPhone 校验中国11位手机号。
// 规则：以1开头，第二位为3-9，共11位纯数字。
func validateCnPhone(fl validator.FieldLevel) bool {
	return cnPhoneRegex.MatchString(fl.Field().String())
}
