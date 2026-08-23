package handler

import (
	"net/http"
	"time"
	"weblog-app/service"

	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(as service.AuthService) *AuthHandler {
	return &AuthHandler{authService: as}
}

func (h *AuthHandler) GetLogin(c echo.Context) error {
	return c.Render(http.StatusOK, "login.html", map[string]interface{}{})
}

func (h *AuthHandler) PostLogin(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	_, token, err := h.authService.Login(username, password)
	if err != nil {
		return c.Render(http.StatusBadRequest, "login.html", map[string]interface{}{
			"Error": err.Error(),
		})
	}

	c.SetCookie(&http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})

	return c.Redirect(http.StatusSeeOther, "/")
}

func (h *AuthHandler) GetSignup(c echo.Context) error {
	return c.Render(http.StatusOK, "signup.html", map[string]interface{}{})
}

func (h *AuthHandler) PostSignup(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	_, token, err := h.authService.Signup(username, password)
	if err != nil {
		return c.Render(http.StatusBadRequest, "signup.html", map[string]interface{}{
			"Error": err.Error(),
		})
	}

	c.SetCookie(&http.Cookie{
		Name:     "session_token",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})

	return c.Redirect(http.StatusSeeOther, "/")
}

func (h *AuthHandler) PostLogout(c echo.Context) error {
	cookie, err := c.Cookie("session_token")
	if err == nil && cookie.Value != "" {
		h.authService.Logout(cookie.Value)
	}

	c.SetCookie(&http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	})

	return c.Redirect(http.StatusSeeOther, "/login")
}