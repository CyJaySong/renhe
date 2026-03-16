package rvalid

import (
	zhLocales "github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

var zhTranslator1 ut.Translator
var zhTranslator2 ut.Translator

// init zhTranslator
func init() {
	zhLocale := zhLocales.New()
	zhTranslator1, _ = ut.New(zhLocale, zhLocale).GetTranslator("zh")
	zhTranslator2, _ = ut.New(zhLocale, zhLocale).GetTranslator("zh")
}

// ZhTranslate 返回中文翻译
func ZhTranslate(fe validator.FieldError) string {
	if msg := fe.Translate(zhTranslator1); msg != fe.Error() {
		return msg
	}
	return fe.Translate(zhTranslator2)
}
