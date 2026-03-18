package custom

import (
	"fmt"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"unicode"
)

// RegisterChineseValidation 向验证器注册 chinese 标签及其中文翻译。
func RegisterChineseValidation(v *validator.Validate, zhTrans ut.Translator) (err error) {
	if err = v.RegisterValidation("chinese", validateChinese); err != nil {
		return
	}
	err = v.RegisterTranslation("chinese", zhTrans,
		func(ut ut.Translator) error {
			return ut.Add("chinese", "{0}必须为有效的中文姓名（{1}）", false)
		},
		func(ut ut.Translator, fe validator.FieldError) string {
			desc := fmt.Sprintf("2-%d个汉字，允许含间隔号", chineseNameMaxLen)
			t, _ := ut.T("chinese", fe.Field(), desc)
			return t
		},
	)
	return
}

const chineseNameMaxLen = 20

// validateChinese 校验中国人姓名。
// 规则：2-20个字符，由汉字组成，允许中间包含间隔号（·）用于少数民族姓名。
// 不允许以间隔号开头或结尾，不允许连续间隔号。
func validateChinese(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	runes := []rune(s)
	n := len(runes)
	if n < 2 || n > chineseNameMaxLen {
		return false
	}

	prevDot := false
	for i, r := range runes {
		if r == '·' || r == '•' {
			// 不允许首尾、不允许连续
			if i == 0 || i == n-1 || prevDot {
				return false
			}
			prevDot = true
			continue
		}
		if !unicode.Is(unicode.Han, r) {
			return false
		}
		prevDot = false
	}
	return true
}
