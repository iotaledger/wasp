package authentication

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/authentication/jwt"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"net"
	"strings"
	"time"

	"github.com/iotaledger/wasp/plugins/accounts"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

func AddAuthentication(webAPI WebAPI, registryProvider registry.Provider, configSectionPath string) {
	var config AuthConfiguration
	parameters.GetStruct(configSectionPath, &config)

	accounts := accounts.GetAccounts()

	switch config.Scheme {
	case AuthBasic:
		addBasicAuth(webAPI, accounts)
	case AuthJWT:
		nodeIdentity, err := registryProvider().GetNodeIdentity()

		if err != nil {
			panic(err)
		}

		privateKey := nodeIdentity.PrivateKey.Bytes()
		jwtAuth := addJWTAuth(webAPI, config.JWTConfig, privateKey, accounts)

		if config.AddRoutes {
			authHandler := &AuthHandler{jwt: jwtAuth, accounts: accounts}
			webAPI.POST(AuthRoute(), authHandler.CrossAPIAuthHandler)
		}

	case AuthIpWhitelist:
		addIPWhiteListauth(webAPI, config.IPWhitelistConfig)
	default:
		panic(fmt.Sprintf("Unknown auth scheme %s", config.Scheme))
	}

	if config.AddRoutes {
		addAuthenticationStatus(webAPI, config)
	}
}

func initJWT(duration int, nodeId string, privateKey []byte, accounts *[]accounts.Account) (*jwt.JWTAuth, func(context echo.Context) bool, jwt.MiddlewareValidator, error) {
	jwtAuth, err := jwt.NewJWTAuth(time.Duration(duration)*time.Hour, nodeId, privateKey)

	if err != nil {
		return nil, nil, nil, err
	}

	jwtAuthSkipper := func(context echo.Context) bool {
		path := context.Request().RequestURI

		if strings.HasSuffix(path, AuthRoute()) || strings.HasSuffix(path, AuthStatusRoute()) || path == "/" {
			return true
		}

		return false
	}

	jwtAuthAllow := func(_ echo.Context, claims *jwt.JWTAuthClaims) bool {
		isValidSubject := false

		for _, account := range *accounts {
			if claims.VerifySubject(account.Username) {
				isValidSubject = true
			}
		}

		if isValidSubject {
			return true
		}

		return false
	}

	return jwtAuth, jwtAuthSkipper, jwtAuthAllow, nil
}

func addJWTAuth(webAPI WebAPI, config JWTAuthConfiguration, privateKey []byte, accounts *[]accounts.Account) *jwt.JWTAuth {
	duration := config.Duration

	if duration <= 0 {
		duration = 24
	}

	jwtAuth, jwtSkipper, jwtAuthAllow, _ := initJWT(duration, "wasp0", privateKey, accounts)

	webAPI.Use(jwtAuth.Middleware(jwtSkipper, jwtAuthAllow))

	return jwtAuth
}

func addBasicAuth(webAPI WebAPI, accounts *[]accounts.Account) {
	webAPI.Use(middleware.BasicAuth(func(user, password string, c echo.Context) (bool, error) {

		for _, v := range *accounts {
			if user == v.Username && password == v.Password {
				return true, nil
			}
		}

		return false, nil
	}))
}

func addIPWhiteListauth(webAPI WebAPI, config IPWhiteListAuthConfiguration) {
	ipWhiteList := createIPWhiteList(config)
	webAPI.Use(protected(ipWhiteList))
}

func createIPWhiteList(config IPWhiteListAuthConfiguration) []net.IP {
	r := make([]net.IP, 0)
	for _, ip := range config.IPWhiteList {
		r = append(r, net.ParseIP(ip))
	}
	return r
}

func protected(whitelist []net.IP) echo.MiddlewareFunc {
	isAllowed := func(ip net.IP) bool {
		if ip.IsLoopback() {
			return true
		}
		for _, whitelistedIP := range whitelist {
			if ip.Equal(whitelistedIP) {
				return true
			}
		}
		return false
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			parts := strings.Split(c.Request().RemoteAddr, ":")
			if len(parts) == 2 {
				ip := net.ParseIP(parts[0])
				if ip != nil && isAllowed(ip) {
					return next(c)
				}
			}
			log.Printf("Blocking request from %s: %s %s", c.Request().RemoteAddr, c.Request().Method, c.Request().RequestURI)
			return echo.ErrUnauthorized
		}
	}
}
