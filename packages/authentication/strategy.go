package authentication

import (
	"fmt"
	"time"

	"github.com/iotaledger/wasp/packages/authentication/shared"

	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/users"
)

const (
	AuthJWT         = "jwt"
	AuthBasic       = "basic"
	AuthIPWhitelist = "ip"
	AuthNone        = "none"
)

type JWTAuthConfiguration struct {
	Duration time.Duration `default:"24h" usage:"jwt token lifetime"`
}

type BasicAuthConfiguration struct {
	Username string `default:"wasp" usage:"the username which grants access to the service"`
}

type IPWhiteListAuthConfiguration struct {
	Whitelist []string `default:"0.0.0.0" usage:"a list of ips that are allowed to access the service"`
}

type AuthConfiguration struct {
	Scheme string `default:"ip" usage:"selects which authentication to choose"`

	JWTConfig         JWTAuthConfiguration         `name:"jwt" usage:"defines the jwt configuration"`
	BasicAuthConfig   BasicAuthConfiguration       `name:"basic" usage:"defines the basic auth configuration"`
	IPWhitelistConfig IPWhiteListAuthConfiguration `name:"ip" usage:"defines the whitelist configuration"`
}

type WebAPI interface {
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	Use(middleware ...echo.MiddlewareFunc)
}

func AddNoneAuth(webAPI WebAPI) {
	// Adds a middleware to set the authContext to authenticated.
	// All routes will be open to everyone, so use it in private environments only.
	// Handle with care!
	noneFunc := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authContext := c.Get("auth").(*AuthContext)

			authContext.isAuthenticated = true

			return next(c)
		}
	}

	webAPI.Use(noneFunc)
}

func AddAuthentication(webAPI WebAPI, registryProvider registry.Provider, authConfig AuthConfiguration, claimValidator ClaimValidator) {
	userMap := users.All()

	addAuthContext(webAPI, authConfig)

	switch authConfig.Scheme {
	case AuthBasic:
		AddBasicAuth(webAPI, userMap)
	case AuthJWT:
		nodeIdentity := registryProvider().GetNodeIdentity()

		privateKey := nodeIdentity.GetPrivateKey().AsBytes()

		// The primary claim is the one mandatory claim that gives access to api/webapi/alike
		jwtAuth := AddJWTAuth(webAPI, authConfig.JWTConfig, privateKey, userMap, claimValidator)

		authHandler := &AuthHandler{Jwt: jwtAuth, Users: userMap}
		webAPI.POST(shared.AuthRoute(), authHandler.CrossAPIAuthHandler)

	case AuthIPWhitelist:
		AddIPWhiteListAuth(webAPI, authConfig.IPWhitelistConfig)

	case AuthNone:
		AddNoneAuth(webAPI)

	default:
		panic(fmt.Sprintf("Unknown auth scheme %s", authConfig.Scheme))
	}

	addAuthenticationStatus(webAPI, authConfig)
}

func addAuthContext(webAPI WebAPI, config AuthConfiguration) {
	webAPI.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &AuthContext{
				scheme: config.Scheme,
			}

			c.Set("auth", cc)

			return next(c)
		}
	})
}
