package rhttp

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

// WriteBizResJSON 可选中间件：控制器注册的 handler 成功返回 *BizRes 后，将其原样 JSON 写出。
// 不包统一 envelope（code/message/data）；业务若需 envelope 请自行封装或另写中间件。
// 若响应已提交或无 BizRes，则不改写。
func WriteBizResJSON() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			err := next(c)
			if err != nil {
				return err
			}
			// 已写过响应则不覆盖
			if resp, uerr := echo.UnwrapResponse(c.Response()); uerr == nil && resp.Committed {
				return nil
			}
			res := GetCtrlFuncBizRes(c)
			if res == nil {
				return nil
			}
			return c.JSON(http.StatusOK, res)
		}
	}
}
