package custom

import (
	emoji "github.com/Andrew-M-C/go.emoji"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

// RegisterEmojiValidation 向验证器注册 noemoji 标签及其中文翻译。
func RegisterEmojiValidation(v *validator.Validate, zhTrans ut.Translator) (err error) {
	if err = v.RegisterValidation("noemoji", validateNoEmoji); err != nil {
		return
	}
	err = v.RegisterTranslation("noemoji", zhTrans,
		func(ut ut.Translator) error {
			return ut.Add("noemoji", "{0}不能包含表情符号", false)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("noemoji", fe.Field())
			return t
		},
	)
	return
}

// validateNoEmoji 校验字符串不包含 emoji。
func validateNoEmoji(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	if len(s) == 0 {
		return true
	}
	it := emoji.IterateChars(s)
	for it.Next() {
		if it.CurrentIsEmoji() {
			return false
		}
	}
	return true
}
