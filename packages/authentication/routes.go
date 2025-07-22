// Package authentication implements authentication mechanisms and routes for secure API access.
package authentication

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/iotaledger/wasp/v2/packages/authentication/shared"
	"github.com/iotaledger/wasp/v2/packages/registry"
	"github.com/iotaledger/wasp/v2/packages/users"
	"github.com/iotaledger/wasp/v2/packages/webapi/interfaces"
)

const (
	AuthNone = "none"
	AuthJWT  = "jwt"
)

type JWTAuthConfiguration struct {
	Duration time.Duration `default:"24h" usage:"jwt token lifetime"`
}

type AuthConfiguration struct {
	Scheme string `default:"ip" usage:"selects which authentication to choose"`

	JWTConfig JWTAuthConfiguration `name:"jwt" usage:"defines the jwt configuration"`
}

type WebAPI interface {
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	Use(middleware ...echo.MiddlewareFunc)
}

func AddAuthentication(
	apiRoot echoswagger.ApiRoot,
	userManager *users.UserManager,
	nodeIdentityProvider registry.NodeIdentityProvider,
	authConfig AuthConfiguration,
	mocker interfaces.Mocker,
) echo.MiddlewareFunc {
	echoRoot := apiRoot.Echo()
	authGroup := apiRoot.Group("auth", "")

	// initialize AuthContext obj as var in echo.Context
	echoRoot.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("auth", &AuthContext{
				scheme: authConfig.Scheme,
			})

			return next(c)
		}
	})

	// set AuthInfo route
	authGroup.GET(shared.AuthInfoRoute(), authInfoHandler(authConfig)).
		AddResponse(http.StatusOK, "Login was successful", mocker.Get(shared.AuthInfoModel{}), nil).
		SetOperationId("authInfo").
		SetSummary("Get information about the current authentication mode")

	// set Auth route
	var middleware echo.MiddlewareFunc
	var handler echo.HandlerFunc
	switch authConfig.Scheme {
	case AuthJWT:
		var jwtAuth *JWTAuth
		nodeIDKeypair := nodeIdentityProvider.NodeIdentity()

		// The primary claim is the one mandatory claim that gives access to api/webapi/alike
		jwtAuth, middleware = GetJWTAuthMiddleware(authConfig.JWTConfig, nodeIDKeypair, userManager)
		authHandler := &AuthHandler{Jwt: jwtAuth, UserManager: userManager}
		handler = authHandler.JWTLoginHandler

	case AuthNone:
		middleware = GetNoneAuthMiddleware()
		handler = nil

	default:
		panic(fmt.Sprintf("Unknown auth scheme %s", authConfig.Scheme))
	}

	authGroup.POST(shared.AuthRoute(), handler).
		AddParamBody(mocker.Get(shared.LoginRequest{}), "", "The login request", true).
		AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", nil, nil).
		AddResponse(http.StatusMethodNotAllowed, "auth type: none", nil, nil).
		AddResponse(http.StatusOK, "Login was successful", mocker.Get(shared.LoginResponse{}), nil).
		SetOperationId("authenticate").
		SetSummary("Authenticate towards the node")
	return middleware
}

func authInfoHandler(authConfig AuthConfiguration) func(c echo.Context) error {
	return func(c echo.Context) error {
		model := shared.AuthInfoModel{
			Scheme: authConfig.Scheme,
		}

		if model.Scheme == AuthJWT {
			model.AuthURL = shared.AuthRoute()
		}

		return c.JSON(http.StatusOK, model)
	}
}
