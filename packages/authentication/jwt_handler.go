package authentication

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"time"

	"github.com/iotaledger/wasp/packages/authentication/shared"

	"github.com/iotaledger/wasp/packages/users"
	"github.com/labstack/echo/v4"
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

func (a *AuthHandler) GetTypedClaims(user *users.UserData) (*WaspClaims, error) {
	claims := WaspClaims{}
	fakeClaims := make(map[string]interface{})

	for _, v := range user.Claims {
		fakeClaims[v] = true
	}

	// TODO: Find a better solution for
	// Turning a list of strings into WaspClaims map by their json tag names
	enc, err := json.Marshal(fakeClaims)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(enc, &claims)

	if err != nil {
		return nil, err
	}

	return &claims, err
}

func (a *AuthHandler) CrossAPIAuthHandler(c echo.Context) error {
	request := &shared.LoginRequest{}

	if err := c.Bind(request); err != nil {
		return c.NoContent(http.StatusUnauthorized)
	}

	if !a.validateLogin(request.Username, request.Password) {
		return c.NoContent(http.StatusUnauthorized)
	}

	user := users.GetUserByName(request.Username)

	if user == nil {
		return c.NoContent(http.StatusUnauthorized)
	}

	claims, err := a.GetTypedClaims(user)
	if err != nil {
		return c.NoContent(http.StatusUnauthorized)
	}

	token, err := a.Jwt.IssueJWT(request.Username, claims)
	if err != nil {
		return c.NoContent(http.StatusUnauthorized)
	}

	contentType := c.Request().Header.Get(echo.HeaderContentType)

	if contentType == echo.MIMEApplicationJSON {
		return c.JSON(http.StatusOK, shared.LoginResponse{JWT: token})
	}

	if contentType == echo.MIMEApplicationForm {
		cookie := http.Cookie{
			Name:     "jwt",
			Value:    token,
			HttpOnly: true, // JWT Token will be stored in a http only cookie, this is important to mitigate XSS/XSRF attacks
			Expires:  time.Now().Add(a.Jwt.durationHours),
			Path:     "/",
			SameSite: http.SameSiteStrictMode,
		}

		c.SetCookie(&cookie)

		return c.Redirect(http.StatusMovedPermanently, shared.AuthRouteSuccess())
	}

	return c.NoContent(http.StatusUnauthorized)
}
