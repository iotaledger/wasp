package authentication

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/packages/authentication/shared"
	"github.com/iotaledger/wasp/packages/authentication/shared/permissions"
	"github.com/iotaledger/wasp/packages/users"
)

// Errors
var (
	ErrJWTInvalidClaims = echo.NewHTTPError(http.StatusUnauthorized, "invalid jwt claims")
	ErrInvalidJWT       = echo.NewHTTPError(http.StatusUnauthorized, "token is invalid")
)

const (
	JWTContextKey = "jwt"
)

type JWTAuth struct {
	duration time.Duration
	nodeID   string
	secret   []byte
}

type MiddlewareValidator = func(c echo.Context, authContext *AuthContext) bool

func NewJWTAuth(duration time.Duration, nodeID string, secret []byte) *JWTAuth {
	return &JWTAuth{
		duration: duration,
		nodeID:   nodeID,
		secret:   secret,
	}
}

type WaspClaims struct {
	jwt.RegisteredClaims
	Permissions map[string]struct{} `json:"permissions"`
}

func (c *WaspClaims) HasPermission(permission string) bool {
	_, exists := c.Permissions[permission]

	if exists {
		return true
	}

	if permission == permissions.Read {
		// If a user only has write permissions, it should still be able to read.
		_, exists = c.Permissions[permissions.Write]

		return exists
	}

	return false
}

func (c *WaspClaims) compare(field, expected string) bool {
	if field == "" {
		return false
	}
	if subtle.ConstantTimeCompare([]byte(field), []byte(expected)) != 0 {
		return true
	}

	return false
}

func (c *WaspClaims) VerifySubject(expected string) bool {
	return c.compare(c.Subject, expected)
}

func (j *JWTAuth) IssueJWT(username string, authClaims *WaspClaims) (string, error) {
	now := time.Now()

	// Set claims
	registeredClaims := jwt.RegisteredClaims{
		Subject:   username,
		Issuer:    j.nodeID,
		Audience:  jwt.ClaimStrings{j.nodeID},
		ID:        fmt.Sprintf("%d", now.Unix()),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
	}

	if j.duration > 0 {
		registeredClaims.ExpiresAt = jwt.NewNumericDate(now.Add(j.duration))
	}

	authClaims.RegisteredClaims = registeredClaims

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, authClaims)

	// Generate encoded token and send it as response.
	return token.SignedString(j.secret)
}

var DefaultJWTDuration time.Duration

func AddJWTAuth(config JWTAuthConfiguration, privateKey []byte, userManager *users.UserManager, claimValidator ClaimValidator) (*JWTAuth, func() echo.MiddlewareFunc) {
	duration := config.Duration

	// If durationHours is 0, we set 24h as the default duration
	if duration == 0 {
		duration = DefaultJWTDuration
	}

	// FIXME: replace "wasp" as nodeID
	jwtAuth := NewJWTAuth(duration, "wasp", privateKey)

	authMiddleware := func() echo.MiddlewareFunc {
		return echojwt.WithConfig(echojwt.Config{
			ContextKey: JWTContextKey,
			NewClaimsFunc: func(c echo.Context) jwt.Claims {
				return &WaspClaims{}
			},
			Skipper: func(c echo.Context) bool {
				path := c.Request().URL.Path
				if path == "/" ||
					strings.HasSuffix(path, shared.AuthRoute()) ||
					strings.HasSuffix(path, shared.AuthInfoRoute()) ||
					strings.HasPrefix(path, "/doc") {
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

				audience, err := claims.GetAudience()
				if err != nil {
					return nil, err
				}
				b, err := audience.MarshalJSON()
				if err != nil {
					return nil, err
				}
				if subtle.ConstantTimeCompare(b, []byte(fmt.Sprintf("[%q]", jwtAuth.nodeID))) == 0 {
					return nil, fmt.Errorf("not in audience")
				}

				userMap := userManager.Users()
				if _, ok := userMap[claims.Subject]; !ok {
					return nil, fmt.Errorf("invalid subject")
				}

				authContext := c.Get("auth").(*AuthContext)
				authContext.isAuthenticated = true
				authContext.claims = claims

				return token, nil
			},
		})
	}

	return jwtAuth, authMiddleware
}
