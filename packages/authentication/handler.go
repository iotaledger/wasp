package authentication

import (
	jwt2 "github.com/iotaledger/wasp/packages/authentication/jwt"
	"github.com/iotaledger/wasp/plugins/accounts"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

type AuthHandler struct {
	jwt      *jwt2.JWTAuth
	accounts *[]accounts.Account
}

func (s *AuthHandler) validateLogin(username string, password string) bool {
	for _, v := range *s.accounts {
		if username == v.Username && password == v.Password {
			return true
		}
	}

	return false
}

func (s *AuthHandler) CrossAPIAuthHandler(c echo.Context) error {

	type loginRequest struct {
		JWT      string `json:"jwt"`
		Username string `json:"username" form:"Username"`
		Password string `json:"password" form:"Password"`
	}

	request := &loginRequest{}

	if err := c.Bind(request); err != nil {
		return errors.WithMessage(err, "invalid request, error: %s")
	}

	if !s.validateLogin(request.Username, request.Password) {
		return echo.ErrUnauthorized
	}

	account := accounts.GetAccountByName(request.Username)

	if account == nil {
		return echo.ErrUnauthorized
	}

	claims, err := account.GetTypedClaims()

	if err != nil {
		return err
	}

	token, err := s.jwt.IssueJWT(request.Username, claims)

	if err != nil {
		return err
	}

	contentType := c.Request().Header.Get(echo.HeaderContentType)

	if contentType == echo.MIMEApplicationJSON {
		return c.JSON(http.StatusOK, map[string]string{
			"jwt": token,
		})
	}

	if contentType == echo.MIMEApplicationForm {
		cookie := http.Cookie{
			Name:     "jwt",
			Value:    token,
			HttpOnly: true, // JWT Token will be stored in a http only cookie, this is important to mitigate XSS/XSRF attacks
			Expires:  time.Now().Add(24 * time.Hour),
			Path:     "/",
			SameSite: http.SameSiteStrictMode,
		}

		c.SetCookie(&cookie)

		return c.Redirect(http.StatusOK, "/")
	}

	return c.NoContent(http.StatusUnauthorized)
}
