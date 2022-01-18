package auth

import (
	"fmt"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/mitchellh/mapstructure"
	"net"
	"strings"
	"time"

	jwt "github.com/iotaledger/wasp/packages/jwt_auth"
	"github.com/iotaledger/wasp/plugins/accounts"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/pangpanglabs/echoswagger/v2"
)

const (
	Authentication_JWT   = "jwt"
	Authentication_BASIC = "basic"
	Authentication_IP    = "ip"
)

type BaseAuthConfiguration map[string]interface{}

func (b *BaseAuthConfiguration) GetScheme() string {
	var config AuthConfiguration
	mapstructure.Decode(b, &config)
	return config.Scheme
}

func (b *BaseAuthConfiguration) GetJWTAuthConfiguration() JWTAuthConfiguration {
	var config JWTAuthConfiguration
	mapstructure.Decode(b, &config)
	return config
}

func (b *BaseAuthConfiguration) GetBasicAuthConfiguration() BasicAuthConfiguration {
	var config BasicAuthConfiguration
	mapstructure.Decode(b, &config)
	return config
}

func (b *BaseAuthConfiguration) GetIPWhiteListAuthConfiguration() IPWhiteListAuthConfiguration {
	var config IPWhiteListAuthConfiguration
	mapstructure.Decode(b, &config)
	return config
}

type AuthConfiguration struct {
	Scheme string `mapstructure:"scheme"`
}

type JWTAuthConfiguration struct {
	AuthConfiguration `mapstructure:",squash"`
	Duration          int      `mapstructure:"duration"`
	Claims            []string `mapstructure:"claims"`
}

type BasicAuthConfiguration struct {
	AuthConfiguration `mapstructure:",squash"`
}

type IPWhiteListAuthConfiguration struct {
	AuthConfiguration `mapstructure:",squash"`
	IPWhiteList       []string `mapstructure:"adminWhitelist"`
}

// TODO: Find better alternative for swagger/echo differences

type PostHandler = func(route string, cb echo.HandlerFunc)
type UseHandler = func(middleware ...echo.MiddlewareFunc)

func AddAuthenticationWebAPI(adm echoswagger.ApiGroup, config BaseAuthConfiguration) {
	var postHandler = func(route string, cb echo.HandlerFunc) {
		adm.POST(route, cb)
	}

	addAuthentication(adm.EchoGroup().Use, postHandler, config)
}

func AddAuthenticationDashboard(e *echo.Echo, config BaseAuthConfiguration) {
	var postHandler = func(route string, cb echo.HandlerFunc) {
		e.POST(route, cb)
	}

	addAuthentication(e.Use, postHandler, config)
}

// Usually you would pass echo.Echo, but to keep it generalized for WebAPI/Dashboard pass the Use function directly..
func addAuthentication(use func(middleware ...echo.MiddlewareFunc), post PostHandler, config BaseAuthConfiguration) {
	accounts := accounts.GetAccounts()

	switch config.GetScheme() {
	case Authentication_BASIC:
		addBasicAuth(use, accounts)
	case Authentication_JWT:
		jwtAuth := addJWTAuth(use, config.GetJWTAuthConfiguration(), accounts)

		authHandler := &AuthRoute{jwt: jwtAuth, accounts: accounts}
		post("/adm/auth", authHandler.CrossAPIAuthHandler)
	case Authentication_IP:
		addIPWhiteListauth(use, config.GetIPWhiteListAuthConfiguration())
	default:
		panic(fmt.Sprintf("Unknown auth scheme %s", scheme))
	}
}

func initJWT(duration int, nodeId string, accounts *[]accounts.Account) (*jwt.JWTAuth, func(context echo.Context) bool, jwt.MiddlewareValidator, error) {
	privKey, _, _ := crypto.GenerateKeyPair(2, 256)
	jwtAuth, err := jwt.NewJWTAuth(time.Duration(duration)*time.Hour, nodeId, privKey)

	if err != nil {
		return nil, nil, nil, err
	}

	jwtAuthSkipper := func(context echo.Context) bool {
		path := context.Request().RequestURI

		if strings.HasSuffix(path, "/auth") {
			return true
		}

		return false
	}

	// Only allow JWT created for the dashboard
	jwtAuthAllow := func(_ echo.Context, claims *jwt.AuthClaims) bool {
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

func addJWTAuth(use func(middleware ...echo.MiddlewareFunc), config JWTAuthConfiguration, accounts *[]accounts.Account) *jwt.JWTAuth {
	duration := config.Duration

	if duration <= 0 {
		duration = 24
	}

	jwtAuth, jwtSkipper, jwtAuthAllow, _ := initJWT(duration, "wasp0", accounts)

	use(jwtAuth.Middleware(jwtSkipper, jwtAuthAllow))

	return jwtAuth
}

func addBasicAuth(use func(middleware ...echo.MiddlewareFunc), accounts *[]accounts.Account) {
	use(middleware.BasicAuth(func(user, password string, c echo.Context) (bool, error) {

		for _, v := range *accounts {
			if user == v.Username && password == v.Password {
				return true, nil
			}
		}

		return false, nil
	}))
}

func addIPWhiteListauth(use func(middleware ...echo.MiddlewareFunc), config IPWhiteListAuthConfiguration) {
	ipWhiteList := createIPWhiteList(config)
	use(protected(ipWhiteList))
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
