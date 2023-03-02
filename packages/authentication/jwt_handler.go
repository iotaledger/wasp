package authentication

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/hive.go/web/basicauth"
	"github.com/iotaledger/wasp/packages/authentication/shared"
	"github.com/iotaledger/wasp/packages/users"
)

const headerXForwardedPrefix = "X-Forwarded-Prefix"

type AuthHandler struct {
	Jwt         *JWTAuth
	UserManager *users.UserManager
}

func (a *AuthHandler) validateLogin(user *users.User, password string) bool {
	valid, err := basicauth.VerifyPassword([]byte(password), user.PasswordSalt, user.PasswordHash)
	if err != nil {
		return false
	}

	return valid
}

func (a *AuthHandler) stageAuthRequest(c echo.Context) (string, error) {
	request := &shared.LoginRequest{}

	if err := c.Bind(request); err != nil {
		return "", errors.New("invalid form data")
	}

	user, err := a.UserManager.User(request.Username)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if !a.validateLogin(user, request.Password) {
		return "", errors.New("invalid credentials")
	}

	claims := &WaspClaims{
		Permissions: user.Permissions,
	}

	token, err := a.Jwt.IssueJWT(request.Username, claims)
	if err != nil {
		return "", errors.New("unable to login")
	}

	return token, nil
}

func (a *AuthHandler) handleJSONAuthRequest(c echo.Context, token string, errorResult error) error {
	if errorResult != nil {
		return c.JSON(http.StatusUnauthorized, shared.LoginResponse{Error: errorResult})
	}

	return c.JSON(http.StatusOK, shared.LoginResponse{JWT: token})
}

func (a *AuthHandler) redirect(c echo.Context, uri string) error {
	return c.Redirect(http.StatusFound, c.Request().Header.Get(headerXForwardedPrefix)+uri)
}

func (a *AuthHandler) handleFormAuthRequest(c echo.Context, token string, errorResult error) error {
	if errorResult != nil {
		// TODO: Add sessions to get rid of the query parameter?
		return a.redirect(c, fmt.Sprintf("%s?error=%s", shared.AuthRoute(), errorResult))
	}

	cookie := http.Cookie{
		Name:     "jwt",
		Value:    token,
		HttpOnly: true, // JWT Token will be stored in a http only cookie, this is important to mitigate XSS/XSRF attacks
		Expires:  time.Now().Add(a.Jwt.duration),
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	}

	c.SetCookie(&cookie)

	return a.redirect(c, shared.AuthRouteSuccess())
}

func (a *AuthHandler) CrossAPIAuthHandler(c echo.Context) error {
	token, errorResult := a.stageAuthRequest(c)

	contentType := c.Request().Header.Get(echo.HeaderContentType)

	if contentType == echo.MIMEApplicationJSON {
		return a.handleJSONAuthRequest(c, token, errorResult)
	}

	if contentType == echo.MIMEApplicationForm {
		return a.handleFormAuthRequest(c, token, errorResult)
	}

	return errors.New("invalid login request")
}
