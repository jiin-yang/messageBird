package middleware

import "github.com/labstack/echo/v4"

func CommonHeaderSetterMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderAccept, "application/json")
		c.Response().Header().Set(echo.HeaderContentType, "application/json; charset=UTF-8")
		return next(c)
	}
}
