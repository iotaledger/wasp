package authentication

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"time"

	"github.com/iotaledger/wasp/packages/users"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
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
	type loginRequest struct {
		JWT      string `json:"jwt"`
		Username string `json:"username" form:"Username"`
		Password string `json:"password" form:"Password"`
	}

	request := &loginRequest{}

	if err := c.Bind(request); err != nil {
		return errors.WithMessage(err, "invalid request, error: %a")
	}

	if !a.validateLogin(request.Username, request.Password) {
		return echo.ErrUnauthorized
	}

	user := users.GetUserByName(request.Username)

	if user == nil {
		return echo.ErrUnauthorized
	}

	claims, err := a.GetTypedClaims(user)
	if err != nil {
		return err
	}

	token, err := a.Jwt.IssueJWT(request.Username, claims)
	if err != nil {
		return err
	}

	contentType := c.Request().Header.Get(echo.HeaderContentType)

	if contentType == echo.MIMEApplicationJSON {
		return c.JSON(http.StatusMovedPermanently, map[string]string{
			"jwt": token,
		})
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

		return c.Redirect(http.StatusMovedPermanently, AuthRouteSuccess())
	}

	return c.NoContent(http.StatusUnauthorized)
}
