package authentication

import (
	"fmt"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/users"
	"github.com/labstack/echo/v4"
)

const (
	AuthJWT         = "jwt"
	AuthBasic       = "basic"
	AuthIpWhitelist = "ip"
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
	UserName string `koanf:"userName"`
}

type IPWhiteListAuthConfiguration struct {
	IPWhiteList []string `koanf:"whitelist"`
}

type WebAPI interface {
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	Use(middleware ...echo.MiddlewareFunc)
}

var log = logger.NewLogger("authentication")

func AddAuthentication(webAPI WebAPI, registryProvider registry.Provider, configSectionPath string, claimValidator ClaimValidator) {
	var config AuthConfiguration
	parameters.GetStruct(configSectionPath, &config)

	users := users.All()

	addAuthContext(webAPI, config)

	switch config.Scheme {
	case AuthBasic:
		AddBasicAuth(webAPI, users)
	case AuthJWT:
		nodeIdentity, err := registryProvider().GetNodeIdentity()

		if err != nil {
			panic(err)
		}

		privateKey := nodeIdentity.PrivateKey.Bytes()

		// The primary claim is the one mandatory claim that gives access to api/webapi/alike
		jwtAuth := AddJWTAuth(webAPI, config.JWTConfig, privateKey, users, claimValidator)

		if config.AddRoutes {
			authHandler := &AuthHandler{Jwt: jwtAuth, Users: users}
			webAPI.POST(AuthRoute(), authHandler.CrossAPIAuthHandler)
		}

	case AuthIpWhitelist:
		AddIPWhiteListAuth(webAPI, config.IPWhitelistConfig)
	default:
		panic(fmt.Sprintf("Unknown auth scheme %s", config.Scheme))
	}

	if config.AddRoutes {
		addAuthenticationStatus(webAPI, config)
	}
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
