package auth

import (
	"github.com/iotaledger/wasp/plugins/accounts"
	"net/http"

	"github.com/iotaledger/wasp/packages/jwt_auth"
	"github.com/pkg/errors"

	"github.com/labstack/echo/v4"
)

type AuthRoute struct {
	jwt      *jwt_auth.JWTAuth
	accounts *[]accounts.Account
}

func (s *AuthRoute) validateLogin(username string, password string) bool {
	for _, v := range *s.accounts {
		if username == v.Username && password == v.Password {
			return true
		}
	}

	return false
}

func (s *AuthRoute) CrossAPIAuthHandler(c echo.Context) error {

	type loginRequest struct {
		JWT      string `json:"jwt"`
		User     string `json:"user"`
		Password string `json:"password"`
	}

	request := &loginRequest{}

	if err := c.Bind(request); err != nil {
		return errors.WithMessage(err, "invalid request, error: %s")
	}

	if !s.validateLogin(request.User, request.Password) {
		return echo.ErrUnauthorized
	}

	account := accounts.GetAccountByName(request.User)
	claims, err := account.GetTypedClaims()

	if err != nil {
		return err
	}

	t, err := s.jwt.IssueJWT(request.User, claims)

	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]string{
		"jwt": t,
	})

	return c.String(http.StatusOK, "Auth stuff")
}
