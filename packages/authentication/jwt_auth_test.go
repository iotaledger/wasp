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

	"github.com/iotaledger/wasp/v2/packages/authentication"
	"github.com/iotaledger/wasp/v2/packages/authentication/shared"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/users"
)

func TestGetJWTAuthMiddleware(t *testing.T) {
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

		nodeIDKeypair := cryptolib.KeyPairFromSeed(cryptolib.SeedFromBytes([]byte("abc")))

		_, middleware := authentication.GetJWTAuthMiddleware(
			authentication.JWTAuthConfiguration{},
			nodeIDKeypair,
			userManager,
		)
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				cc := &authentication.AuthContext{}
				c.Set("auth", cc)
				return next(c)
			}
		})
		e.Use(middleware)

		req := httptest.NewRequest(http.MethodGet, "/test-route", http.NoBody)
		req.Header.Set(echo.HeaderAuthorization, "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIweGNjYTUzMmNmN2RjNWNhNGExNmJiZjE5OTM5ZThiODlkMDMzN2FhNTk5ZDVjOGQxZGY4MDdlNDM4ZjA3MjExOTEiLCJzdWIiOiJ3YXNwIiwiYXVkIjpbIndhc3AiXSwiZXhwIjo2ODI0NjUwMzMyLCJuYmYiOjE2OTI1OTEyODksImlhdCI6MTY5MjU5MTI4OSwianRpIjoiMTY5MjU5MTI4OSIsInBlcm1pc3Npb25zIjp7IndyaXRlIjp7fX19.nFXeqX4i6K7Jmt3nEdaqJXYp2sp35an4EXdz-U5mWtQ")
		res := httptest.NewRecorder()

		e.ServeHTTP(res, req)

		require.Equal(t, http.StatusOK, res.Code)
		require.Equal(t, "{\"iss\":\"0xcca532cf7dc5ca4a16bbf19939e8b89d0337aa599d5c8d1df807e438f0721191\",\"sub\":\"wasp\",\"aud\":[\"wasp\"],\"exp\":6824650332,\"nbf\":1692591289,\"iat\":1692591289,\"jti\":\"1692591289\",\"permissions\":{\"write\":{}}}\n", res.Body.String())
	})

	t.Run("skip", func(t *testing.T) {
		e := echo.New()
		testRootURL := "http://fake-root"
		skipPaths := []string{
			"/",
			shared.AuthRoute(),
			shared.AuthInfoRoute(),
			"/doc",
		}
		notSkipPaths := []string{
			"/aa/",
			"/user/" + shared.AuthRoute(),
			"/bb/doc",
		}
		for _, path := range skipPaths {
			e.GET(path, func(c echo.Context) error {
				_, ok := c.Get(authentication.JWTContextKey).(*jwt.Token)
				require.False(t, ok)
				return c.JSON(http.StatusOK, "")
			})
		}

		_, middleware := authentication.GetJWTAuthMiddleware(
			authentication.JWTAuthConfiguration{},
			cryptolib.NewKeyPair(),
			&users.UserManager{},
		)
		e.Use(middleware)

		for _, path := range skipPaths {
			req := httptest.NewRequest(http.MethodGet, testRootURL+path, http.NoBody)
			res := httptest.NewRecorder()

			e.ServeHTTP(res, req)

			require.Equal(t, http.StatusOK, res.Code)
			require.Equal(t, "\"\"\n", res.Body.String())
		}

		for _, path := range notSkipPaths {
			req := httptest.NewRequest(http.MethodGet, testRootURL+path, http.NoBody)
			req.Header.Set(echo.HeaderAuthorization, "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiIweGNjYTUzMmNmN2RjNWNhNGExNmJiZjE5OTM5ZThiODlkMDMzN2FhNTk5ZDVjOGQxZGY4MDdlNDM4ZjA3MjExOTEiLCJzdWIiOiJ3YXNwIiwiYXVkIjpbIndhc3AiXSwiZXhwIjo2ODI0NjUwMzMyLCJuYmYiOjE2OTI1OTEyODksImlhdCI6MTY5MjU5MTI4OSwianRpIjoiMTY5MjU5MTI4OSIsInBlcm1pc3Npb25zIjp7IndyaXRlIjp7fX19.nFXeqX4i6K7Jmt3nEdaqJXYp2sp35an4EXdz-U5mWtQ")
			res := httptest.NewRecorder()

			e.ServeHTTP(res, req)

			require.Equal(t, http.StatusUnauthorized, res.Code)
		}
	})
}

func TestJWTAuthIssueAndVerify(t *testing.T) {
	e := echo.New()
	e.GET("/test-route", func(c echo.Context) error {
		token := c.Get(authentication.JWTContextKey).(*jwt.Token)
		return c.JSON(http.StatusOK, token.Claims)
	})

	duration := 20 * time.Hour
	username := "wasp"
	nodeIDKeypair := cryptolib.KeyPairFromSeed(cryptolib.SeedFromBytes([]byte("abc")))
	jwtAuth := authentication.NewJWTAuth(duration, nodeIDKeypair)

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

	_, middleware := authentication.GetJWTAuthMiddleware(
		authentication.JWTAuthConfiguration{Duration: duration},
		nodeIDKeypair,
		userManager,
	)
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &authentication.AuthContext{}
			c.Set("auth", cc)
			return next(c)
		}
	})
	e.Use(middleware)

	req := httptest.NewRequest(http.MethodGet, "/test-route", http.NoBody)
	req.Header.Set(echo.HeaderAuthorization, fmt.Sprintf("Bearer %s", jwtString))
	res := httptest.NewRecorder()

	e.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
}
