package auth

import (
	"fmt"
	"strings"
	"time"

	jwt "github.com/iotaledger/wasp/packages/jwt_auth"
	"github.com/iotaledger/wasp/plugins/accounts"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/pangpanglabs/echoswagger/v2"
)

type AuthConfiguration struct {
	Scheme   string
	Duration int
	Claim    string
}

// TODO: Find better alternative for swagger/echo differences

type PostHandler = func(route string, cb echo.HandlerFunc)
type UseHandler = func(middleware ...echo.MiddlewareFunc)

func AddAuthenticationWebAPI(adm echoswagger.ApiGroup, config AuthConfiguration) {
	var postHandler = func(route string, cb echo.HandlerFunc) {
		adm.POST(route, cb)
	}

	addAuthentication(adm.EchoGroup().Use, postHandler, config)
}

func AddAuthenticationDashboard(e *echo.Echo, config AuthConfiguration) {
	var postHandler = func(route string, cb echo.HandlerFunc) {
		e.POST(route, cb)
	}

	addAuthentication(e.Use, postHandler, config)
}

// Usually you would pass echo.Echo, but to keep it generalized for WebAPI/Dashboard pass the Use function directly..
func addAuthentication(use func(middleware ...echo.MiddlewareFunc), post PostHandler, config AuthConfiguration) {
	accounts := accounts.GetAccounts()

	scheme := config.Scheme

	switch scheme {
	case "basic":
		addBasicAuth(use, accounts)
	case "jwt":
		jwtAuth := addJWTAuth(use, config, accounts)

		authHandler := &AuthRoute{jwt: jwtAuth, accounts: accounts}
		post("/adm/auth", authHandler.CrossAPIAuthHandler)
	default:
		panic(fmt.Sprintf("Unknown auth scheme %s", scheme))
	}
}

func initJWT(claim string, duration int, nodeId string, accounts *[]accounts.Account) (*jwt.JWTAuth, func(context echo.Context) bool, jwt.MiddlewareValidator, error) {
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

		fmt.Printf("%v", path)

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

		if isValidSubject && claims.HasClaim(claim) {
			return true
		}

		return false
	}

	return jwtAuth, jwtAuthSkipper, jwtAuthAllow, nil
}

func addJWTAuth(use func(middleware ...echo.MiddlewareFunc), config AuthConfiguration, accounts *[]accounts.Account) *jwt.JWTAuth {
	duration := config.Duration

	if duration <= 0 {
		duration = 24
	}

	jwtAuth, jwtSkipper, jwtAuthAllow, _ := initJWT(config.Claim, duration, "wasp0", accounts)

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
