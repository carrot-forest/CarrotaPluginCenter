package middleware

import (
	"carrota-plugin-center/controllers"
	"carrota-plugin-center/controllers/auth"

	"github.com/labstack/echo/v4"
)

func TokenVerificationMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, err := auth.GetClaimsFromHeader(c)
		if err != nil {
			return controllers.ResponseUnauthorized(c, "Invalid bearer token in header.", err)
		}
		if claims.Valid() != nil {
			return controllers.ResponseUnauthorized(c, "Invalid jwt token.", claims.Valid())
		}

		// 本项目无需校验 token 过期时间
		// if claims.ExpiresAt < time.Now().Unix() {
		// 	return controllers.ResponseUnauthorized(c, "Token expired.", nil)
		// }

		return next(c)
	}
}
