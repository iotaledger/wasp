package authentication

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/v2/packages/authentication/shared"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/users"
)

var DefaultJWTDuration time.Duration

func GetJWTAuthMiddleware(
	config JWTAuthConfiguration,
	nodeIDKeypair *cryptolib.KeyPair,
	userManager *users.UserManager,
) (*JWTAuth, echo.MiddlewareFunc) {
	duration := config.Duration
	// If durationHours is 0, we set 24h as the default duration
	if duration == 0 {
		duration = DefaultJWTDuration
	}

	jwtAuth := NewJWTAuth(duration, nodeIDKeypair)

	authMiddleware := echojwt.WithConfig(echojwt.Config{
		ContextKey: JWTContextKey,
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return &WaspClaims{}
		},
		Skipper: func(c echo.Context) bool {
			path := c.Request().URL.Path
			if path == "/" ||
				path == shared.AuthRoute() ||
				path == shared.AuthInfoRoute() ||
				path == "/doc" {
				return true
			}

			return false
		},
		SigningKey:  jwtAuth.secret,
		TokenLookup: "header:Authorization:Bearer ,cookie:jwt",
		ParseTokenFunc: func(c echo.Context, auth string) (interface{}, error) {
			keyFunc := func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}

				return jwtAuth.secret, nil
			}

			token, err := jwt.ParseWithClaims(
				auth,
				&WaspClaims{},
				keyFunc,
				jwt.WithValidMethods([]string{"HS256"}),
			)
			if err != nil {
				return nil, err
			}
			if !token.Valid {
				return nil, fmt.Errorf("invalid token")
			}

			claims, ok := token.Claims.(*WaspClaims)
			if !ok {
				return nil, fmt.Errorf("wrong JWT claim type")
			}

			userMap := userManager.Users()
			audience, err := claims.GetAudience()
			if err != nil {
				return nil, err
			}
			if _, ok := userMap[audience[0]]; !ok {
				return nil, fmt.Errorf("not in audience")
			}

			if _, ok := userMap[claims.Subject]; !ok {
				return nil, fmt.Errorf("invalid subject")
			}

			authContext := c.Get("auth").(*AuthContext)
			authContext.claims = claims

			return token, nil
		},
	})

	return jwtAuth, authMiddleware
}

func GetNoneAuthMiddleware() echo.MiddlewareFunc {
	// Adds a middleware to set the authContext to authenticated.
	// All routes will be open to everyone, so use it in private environments only.
	// Handle with care!
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			return next(c)
		}
	}
}
