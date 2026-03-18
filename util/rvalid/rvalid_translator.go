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

// ZhTranslate 返回单个字段错误的中文翻译。
// 优先使用 v1 translator，若未命中则回退到 v2 translator。
func ZhTranslate(fe validator.FieldError) string {
	if msg := fe.Translate(zhTranslator1); msg != fe.Error() {
		return msg
	}
	return fe.Translate(zhTranslator2)
}

// TranslateAll 将验证错误批量翻译为中文，返回 字段名→中文消息 的映射。
// 若 err 不是 ValidationErrors 类型，返回 nil。
func TranslateAll(err error) map[string]string {
	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		return nil
	}
	result := make(map[string]string, len(errs))
	for _, fe := range errs {
		result[fe.Field()] = ZhTranslate(fe)
	}
	return result
}

// FirstError 从验证错误中提取第一条中文错误消息。
// 若 err 为 nil 或不是 ValidationErrors 类型，返回空字符串。
func FirstError(err error) string {
	if err == nil {
		return ""
	}
	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		return err.Error()
	}
	if len(errs) == 0 {
		return ""
	}
	return ZhTranslate(errs[0])
}
