package custom

import (
	zhLocales "github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/translations/zh"
	"sync"
)

var (
	validate          *validator.Validate
	zhTranslator      ut.Translator
	onceInitValidator sync.Once
)

func initValidator() {
	onceInitValidator.Do(func() {
		validate = validator.New(validator.WithRequiredStructEnabled())
		validate.SetTagName("v")

		zhLocale := zhLocales.New()
		zhTranslator, _ = ut.New(zhLocale, zhLocale).GetTranslator("zh")
		_ = zh.RegisterDefaultTranslations(validate, zhTranslator)
	})
}
