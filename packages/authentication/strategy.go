package authentication

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/plugins/accounts"
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
	Duration    int    `koanf:"duration"`
	AccessClaim string `koanf:"accessClaim"`
}

type BasicAuthConfiguration struct {
	AccountName string `koanf:"accountName"`
}

type IPWhiteListAuthConfiguration struct {
	IPWhiteList []string `koanf:"whitelist"`
}

type WebAPI interface {
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	Use(middleware ...echo.MiddlewareFunc)
}

func AddAuthentication(webAPI WebAPI, registryProvider registry.Provider, configSectionPath string, primaryClaim string) {
	var config AuthConfiguration
	parameters.GetStruct(configSectionPath, &config)

	accounts := accounts.GetAccounts()

	addAuthContext(webAPI, config)

	switch config.Scheme {
	case AuthBasic:
		AddBasicAuth(webAPI, accounts)
	case AuthJWT:
		nodeIdentity, err := registryProvider().GetNodeIdentity()

		if err != nil {
			panic(err)
		}

		privateKey := nodeIdentity.PrivateKey.Bytes()

		// The primary claim is the one mandatory claim that gives access to api/webapi/alike
		jwtAuth := AddJWTAuth(webAPI, config.JWTConfig, privateKey, accounts, primaryClaim)

		if config.AddRoutes {
			authHandler := &AuthHandler{Jwt: jwtAuth, Accounts: accounts}
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
				Scheme: config.Scheme,
			}

			c.Set("auth", cc)

			return next(c)
		}
	})
}
