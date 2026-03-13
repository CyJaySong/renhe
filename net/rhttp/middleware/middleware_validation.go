package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/cyjaysong/renhe/util/rvalid"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// ValidationMiddleware 拦截 handler 返回的错误，
// 如果是 validator.ValidationErrors，则翻译为中文并返回 400。
// 优先使用 vm tag 的 Message()，无自定义消息时使用 zh 翻译器兜底。
func ValidationMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			if err = next(c); err == nil {
				return
			}
			if ve, ok := errors.AsType[validator.ValidationErrors](err); ok {
				msgs := make([]string, 0, len(ve))
				zhTrans := rvalid.ZhTranslator()
				for _, fe := range ve {
					if msg := fe.Message(); len(msg) > 0 {
						msgs = append(msgs, msg)
					} else if zhTrans != nil {
						msgs = append(msgs, fe.Translate(zhTrans))
					} else {
						msgs = append(msgs, fe.Error())
					}
				}
				return echo.NewHTTPError(http.StatusBadRequest, strings.Join(msgs, ";"))
			}
			return
		}
	}
}
