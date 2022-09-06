package authentication

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/xerrors"

	"github.com/iotaledger/wasp/packages/authentication/shared"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/users"
)

type AuthHandler struct {
	Jwt   *JWTAuth
	Users map[string]*users.UserData
}

func (a *AuthHandler) validateLogin(username, password string) bool {
	userDetail := a.Users[username]

	if userDetail == nil {
		return false
	}

	if subtle.ConstantTimeCompare([]byte(userDetail.Password), []byte(password)) != 0 {
		return true
	}

	return false
}

func (a *AuthHandler) stageAuthRequest(c echo.Context) (string, error) {
	request := &shared.LoginRequest{}

	if err := c.Bind(request); err != nil {
		return "", xerrors.Errorf("Invalid form data")
	}

	if !a.validateLogin(request.Username, request.Password) {
		return "", xerrors.Errorf("Invalid credentials")
	}

	user := users.GetUserByName(request.Username)

	if user == nil {
		return "", xerrors.Errorf("Invalid credentials")
	}

	claims := &WaspClaims{
		Permissions: users.GetPermissionMap(request.Username),
	}

	token, err := a.Jwt.IssueJWT(request.Username, claims)
	if err != nil {
		return "", xerrors.Errorf("Unable to login")
	}

	return token, nil
}

func (a *AuthHandler) handleJSONAuthRequest(c echo.Context, token string, errorResult error) error {
	if errorResult != nil {
		return c.JSON(http.StatusUnauthorized, shared.LoginResponse{Error: errorResult})
	}

	return c.JSON(http.StatusOK, shared.LoginResponse{JWT: token})
}

func (a *AuthHandler) handleFormAuthRequest(c echo.Context, token string, errorResult error) error {
	if errorResult != nil {
		// TODO: Add sessions to get rid of the query parameter?
		return c.Redirect(http.StatusFound, fmt.Sprintf("%s?error=%s", shared.AuthRoute(), errorResult))
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

	return c.Redirect(http.StatusFound, shared.AuthRouteSuccess())
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

	return xerrors.Errorf("Invalid login request")
}
