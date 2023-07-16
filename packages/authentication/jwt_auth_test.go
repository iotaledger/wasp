package authentication_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/authentication"
	"github.com/iotaledger/wasp/packages/authentication/shared"
	"github.com/iotaledger/wasp/packages/users"
)

func TestAddJWTAuth(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		e := echo.New()
		e.GET("/test-route", func(c echo.Context) error {
			token := c.Get(authentication.JWTContextKey).(*jwt.Token)
			return c.JSON(http.StatusOK, token.Claims)
		})

		userManager := users.NewUserManager((func(users []*users.User) error {
			return nil
		}))

		userManager.AddUser(&users.User{
			Name: "wasp",
		})

		_, middleware := authentication.AddJWTAuth(
			authentication.JWTAuthConfiguration{},
			[]byte("abc"),
			userManager,
			nil, // remove claim validator
		)
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				cc := &authentication.AuthContext{}
				c.Set("auth", cc)
				return next(c)
			}
		})
		e.Use(middleware())

		req := httptest.NewRequest(http.MethodGet, "/test-route", http.NoBody)
		req.Header.Set(echo.HeaderAuthorization, "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJ3YXNwIiwic3ViIjoid2FzcCIsImF1ZCI6WyJ3YXNwIl0sImV4cCI6NDg0NTUwNjQ5MiwibmJmIjoxNjg5ODYxNDM2LCJpYXQiOjE2ODk4NjE0MzYsImp0aSI6IjE2ODk4NjE0MzYiLCJwZXJtaXNzaW9ucyI6eyJ3cml0ZSI6e319fQ.VP--725H3xO2Spz6L9twB6Tsm37a26IXVU87cSqRoOM")
		res := httptest.NewRecorder()

		e.ServeHTTP(res, req)

		require.Equal(t, http.StatusOK, res.Code)
		require.Equal(t, "{\"iss\":\"wasp\",\"sub\":\"wasp\",\"aud\":[\"wasp\"],\"exp\":4845506492,\"nbf\":1689861436,\"iat\":1689861436,\"jti\":\"1689861436\",\"permissions\":{\"write\":{}}}\n", res.Body.String())
	})

	t.Run("skip", func(t *testing.T) {
		e := echo.New()
		skipPaths := []string{
			"/",
			shared.AuthRoute(),
			shared.AuthInfoRoute(),
			"/doc",
		}
		for _, path := range skipPaths {
			e.GET(path, func(c echo.Context) error {
				_, ok := c.Get(authentication.JWTContextKey).(*jwt.Token)
				require.False(t, ok)
				return c.JSON(http.StatusOK, "")
			})
		}

		_, middleware := authentication.AddJWTAuth(
			authentication.JWTAuthConfiguration{},
			[]byte(""),
			&users.UserManager{},
			nil, // remove claim validator
		)
		e.Use(middleware())

		for _, path := range skipPaths {
			req := httptest.NewRequest(http.MethodGet, path, http.NoBody)
			res := httptest.NewRecorder()

			e.ServeHTTP(res, req)

			require.Equal(t, http.StatusOK, res.Code)
			require.Equal(t, "\"\"\n", res.Body.String())
		}
	})
}

func TestJWTAuthIssueAndVerify(t *testing.T) {
	e := echo.New()
	e.GET("/test-route", func(c echo.Context) error {
		token := c.Get(authentication.JWTContextKey).(*jwt.Token)
		return c.JSON(http.StatusOK, token.Claims)
	})

	privateKey := []byte("abc")
	duration := 20 * time.Hour
	username := "wasp"
	jwtAuth := authentication.NewJWTAuth(duration, username, privateKey)

	jwtString, err := jwtAuth.IssueJWT(username, &authentication.WaspClaims{
		Permissions: map[string]struct{}{
			"write": {},
		},
	})
	require.NoError(t, err)

	userManager := users.NewUserManager((func(users []*users.User) error {
		return nil
	}))
	userManager.AddUser(&users.User{
		Name: username,
	})

	_, middleware := authentication.AddJWTAuth(
		authentication.JWTAuthConfiguration{Duration: duration},
		privateKey,
		userManager,
		nil, // remove claim validator
	)
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &authentication.AuthContext{}
			c.Set("auth", cc)
			return next(c)
		}
	})
	e.Use(middleware())

	req := httptest.NewRequest(http.MethodGet, "/test-route", http.NoBody)
	req.Header.Set(echo.HeaderAuthorization, fmt.Sprintf("Bearer %s", jwtString))
	res := httptest.NewRecorder()

	e.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
}
