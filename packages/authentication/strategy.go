package authentication

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pangpanglabs/echoswagger/v2"

	"github.com/iotaledger/wasp/packages/authentication/shared"
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

func AddV1Authentication(
	webAPI WebAPI,
	userManager *users.UserManager,
	nodeIdentityProvider registry.NodeIdentityProvider,
	authConfig AuthConfiguration,
	claimValidator ClaimValidator,
) {
	addAuthContext(webAPI, authConfig)

	switch authConfig.Scheme {
	case AuthBasic:
		AddBasicAuth(webAPI, userManager)
	case AuthJWT:
		nodeIdentity := nodeIdentityProvider.NodeIdentity()
		privateKey := nodeIdentity.GetPrivateKey().AsBytes()

		// The primary claim is the one mandatory claim that gives access to api/webapi/alike
		jwtAuth, authMiddleware := AddJWTAuth(authConfig.JWTConfig, privateKey, userManager, claimValidator)

		authHandler := &AuthHandler{Jwt: jwtAuth, UserManager: userManager}
		webAPI.POST(shared.AuthRoute(), authHandler.CrossAPIAuthHandler)
		webAPI.Use(authMiddleware())

	case AuthIPWhitelist:
		AddIPWhiteListAuth(webAPI, authConfig.IPWhitelistConfig)

	case AuthNone:
		AddNoneAuth(webAPI)

	default:
		panic(fmt.Sprintf("Unknown auth scheme %s", authConfig.Scheme))
	}

	addAuthenticationStatus(webAPI, authConfig)
}

// TODO: After deprecating V1 we can slim down this whole strategy handler.
// It is currently needed as the current authentication scheme does not support echoSwagger,
// which leaves authentication out of the client code generator.
// After v1 gets removed:
// * Get rid off basic/ip auth and only keeping 'none' and 'JWT'
// * Properly document the routes with echoSwagger
// * Keep only one AddAuthentication method

func AddV2Authentication(apiRoot echoswagger.ApiRoot,
	userManager *users.UserManager,
	nodeIdentityProvider registry.NodeIdentityProvider,
	authConfig AuthConfiguration,
	claimValidator ClaimValidator,
) func() echo.MiddlewareFunc {
	echoRoot := apiRoot.Echo()
	authGroup := apiRoot.Group("auth", "")

	addAuthContext(echoRoot, authConfig)

	c := &StatusWebAPIModel{
		config: authConfig,
	}

	authGroup.GET(shared.AuthInfoRoute(), c.handleAuthenticationStatus).
		AddResponse(http.StatusOK, "Login was successful", shared.AuthInfoModel{}, nil).
		SetOperationId("authInfo").
		SetSummary("Get information about the current authentication mode")

	switch authConfig.Scheme {
	case AuthJWT:
		nodeIdentity := nodeIdentityProvider.NodeIdentity()
		privateKey := nodeIdentity.GetPrivateKey().AsBytes()

		// The primary claim is the one mandatory claim that gives access to api/webapi/alike
		jwtAuth, jwtMiddleware := AddJWTAuth(authConfig.JWTConfig, privateKey, userManager, claimValidator)

		authHandler := &AuthHandler{Jwt: jwtAuth, UserManager: userManager}
		authGroup.POST(shared.AuthRoute(), authHandler.CrossAPIAuthHandler).
			AddParamBody(shared.LoginRequest{}, "", "The login request", true).
			AddResponse(http.StatusUnauthorized, "Unauthorized (Wrong permissions, missing token)", nil, nil).
			AddResponse(http.StatusOK, "Login was successful", shared.LoginResponse{}, nil).
			SetOperationId("authenticate").
			SetSummary("Authenticate towards the node")

		return jwtMiddleware

	case AuthNone:
		AddNoneAuth(echoRoot)
		return nil

	default:
		panic(fmt.Sprintf("Unknown auth scheme %s", authConfig.Scheme))
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
