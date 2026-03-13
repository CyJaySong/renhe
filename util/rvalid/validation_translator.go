package rvalid

import (
	zhLocales "github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
)

var zhTranslator ut.Translator

// init zhTranslator
func init() {
	zhLocale := zhLocales.New()
	uni := ut.New(zhLocale, zhLocale)
	zhTranslator, _ = uni.GetTranslator("zh")
}

// ZhTranslator 返回中文翻译器实例。
func ZhTranslator() ut.Translator { return zhTranslator }
