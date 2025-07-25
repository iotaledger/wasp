package authentication

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/hive.go/web/basicauth"
	"github.com/iotaledger/wasp/v2/packages/authentication/shared"
	"github.com/iotaledger/wasp/v2/packages/users"
)

type AuthHandler struct {
	Jwt         *JWTAuth
	UserManager *users.UserManager
}

func (a *AuthHandler) JWTLoginHandler(c echo.Context) error {
	if c.Request().Header.Get(echo.HeaderContentType) != echo.MIMEApplicationJSON {
		return errors.New("invalid login request")
	}

	req, user, err := a.parseAuthRequest(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, shared.LoginResponse{Error: err})
	}

	claims := &WaspClaims{
		Permissions: user.Permissions,
	}
	token, err := a.Jwt.IssueJWT(req.Username, claims)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, shared.LoginResponse{Error: fmt.Errorf("unable to login")})
	}

	return c.JSON(http.StatusOK, shared.LoginResponse{JWT: token})
}

func (a *AuthHandler) parseAuthRequest(c echo.Context) (*shared.LoginRequest, *users.User, error) {
	request := &shared.LoginRequest{}

	if err := c.Bind(request); err != nil {
		return nil, nil, fmt.Errorf("invalid form data")
	}

	user, err := a.UserManager.User(request.Username)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid credentials")
	}

	if !validatePassword(user, request.Password) {
		return nil, nil, fmt.Errorf("invalid credentials")
	}

	return request, user, nil
}

func validatePassword(user *users.User, password string) bool {
	valid, err := basicauth.VerifyPassword([]byte(password), user.PasswordSalt, user.PasswordHash)
	if err != nil {
		return false
	}

	return valid
}
