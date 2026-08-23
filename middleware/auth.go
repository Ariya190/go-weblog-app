package middleware

import (
	"net/http"
	"weblog-app/service"

	"github.com/labstack/echo/v4"
)

func AuthMiddleware(authService service.AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie("session_token")
			if err != nil || cookie.Value == "" {
				return c.Redirect(http.StatusSeeOther, "/login")
			}

			userID, err := authService.ValidateSession(cookie.Value)
			if err != nil {
				return c.Redirect(http.StatusSeeOther, "/login")
			}

			c.Set("user_id", userID)
			return next(c)
		}
	}
}