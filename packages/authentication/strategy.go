package authentication

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/authentication/shared"

	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/users"
	"github.com/labstack/echo/v4"
)

const (
	AuthJWT         = "jwt"
	AuthBasic       = "basic"
	AuthIPWhitelist = "ip"
	AuthNone        = "none"
)

type AuthConfiguration struct {
	Scheme    string `koanf:"scheme"`
	AddRoutes bool   `koanf:"addRoutes"`

	JWTConfig         JWTAuthConfiguration         `koanf:"jwt"`
	BasicAuthConfig   BasicAuthConfiguration       `koanf:"basic"`
	IPWhitelistConfig IPWhiteListAuthConfiguration `koanf:"ip"`
}

type JWTAuthConfiguration struct {
	DurationHours int `koanf:"durationHours"`
}

type BasicAuthConfiguration struct {
	UserName string `koanf:"username"`
}

type IPWhiteListAuthConfiguration struct {
	IPWhiteList []string `koanf:"whitelist"`
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

func AddAuthentication(webAPI WebAPI, registryProvider registry.Provider, configSectionPath string, claimValidator ClaimValidator) {
	var config AuthConfiguration

	if err := parameters.GetStruct(configSectionPath, &config); err != nil {
		return
	}

	userMap := users.All()

	addAuthContext(webAPI, config)

	switch config.Scheme {
	case AuthBasic:
		AddBasicAuth(webAPI, userMap)
	case AuthJWT:
		nodeIdentity, err := registryProvider().GetNodeIdentity()
		if err != nil {
			panic(err)
		}

		privateKey := nodeIdentity.PrivateKey.Bytes()

		// The primary claim is the one mandatory claim that gives access to api/webapi/alike
		jwtAuth := AddJWTAuth(webAPI, config.JWTConfig, privateKey, userMap, claimValidator)

		authHandler := &AuthHandler{Jwt: jwtAuth, Users: userMap}
		webAPI.POST(shared.AuthRoute(), authHandler.CrossAPIAuthHandler)

	case AuthIPWhitelist:
		AddIPWhiteListAuth(webAPI, config.IPWhitelistConfig)

	case AuthNone:
		AddNoneAuth(webAPI)

	default:
		panic(fmt.Sprintf("Unknown auth scheme %s", config.Scheme))
	}

	addAuthenticationStatus(webAPI, config)
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
