package authentication

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/iotaledger/wasp/packages/authentication/shared"

	jwt "github.com/golang-jwt/jwt"
	"github.com/iotaledger/wasp/packages/users"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Errors
var (
	ErrJWTInvalidClaims = echo.NewHTTPError(http.StatusUnauthorized, "invalid jwt claims")
	ErrInvalidJWT       = echo.NewHTTPError(http.StatusUnauthorized, "token is invalid")
)

type JWTAuth struct {
	durationHours time.Duration
	nodeID        string
	secret        []byte
}

type MiddlewareValidator = func(c echo.Context, authContext *AuthContext) bool

func NewJWTAuth(durationHours time.Duration, nodeID string, secret []byte) (*JWTAuth, error) {
	return &JWTAuth{
		durationHours: durationHours,
		nodeID:        nodeID,
		secret:        secret,
	}, nil
}

type WaspClaims struct {
	jwt.StandardClaims
	Dashboard  bool `json:"dashboard"`
	API        bool `json:"api"`
	ChainRead  bool `json:"chain.read"`
	ChainWrite bool `json:"chain.write"`
}

const defaultJwtDurationHours = 24

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

func (j *JWTAuth) Middleware(skipper middleware.Skipper, allow MiddlewareValidator) echo.MiddlewareFunc {
	config := middleware.JWTConfig{
		ContextKey:  "jwt",
		Claims:      &WaspClaims{},
		SigningKey:  j.secret,
		TokenLookup: "header:Authorization,cookie:jwt",
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authContext := c.Get("auth").(*AuthContext)

			// skip unprotected endpoints
			if skipper(c) {
				return next(c)
			}

			// use the default JWT middleware to verify and extract the JWT
			handler := middleware.JWTWithConfig(config)(func(c echo.Context) error {
				return nil
			})

			// run the JWT middleware
			if err := handler(c); err != nil {
				return ErrInvalidJWT
			}

			token := c.Get("jwt").(*jwt.Token)

			// validate the signing method we expect
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return ErrInvalidJWT
			}

			// read the claims set by the JWT middleware on the context
			authContext.claims = token.Claims.(*WaspClaims)

			// do extended authClaims validation
			if !authContext.claims.VerifyAudience(j.nodeID, true) {
				return ErrJWTInvalidClaims
			}

			// validate claims
			if !allow(c, authContext) {
				return ErrJWTInvalidClaims
			}

			authContext.isAuthenticated = true

			// go to the next handler
			return next(c)
		}
	}
}

func (j *JWTAuth) IssueJWT(username string, authClaims *WaspClaims) (string, error) {
	now := time.Now()

	// Set claims
	stdClaims := jwt.StandardClaims{
		Subject:   username,
		Issuer:    j.nodeID,
		Audience:  j.nodeID,
		Id:        fmt.Sprintf("%d", now.Unix()),
		IssuedAt:  now.Unix(),
		NotBefore: now.Unix(),
	}

	if j.durationHours > 0 {
		stdClaims.ExpiresAt = now.Add(j.durationHours).Unix()
	}

	authClaims.StandardClaims = stdClaims

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, authClaims)

	// Generate encoded token and send it as response.
	return token.SignedString(j.secret)
}

func (j *JWTAuth) keyFunc(token *jwt.Token) (interface{}, error) {
	// validate the signing method we expect
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}

	return j.secret, nil
}

func (j *JWTAuth) VerifyJWT(token string, allow ClaimValidator) bool {
	t, err := jwt.ParseWithClaims(token, &WaspClaims{}, j.keyFunc)

	if err != nil || !t.Valid {
		return false
	}

	claims, ok := t.Claims.(*WaspClaims)

	if !ok || !claims.VerifyAudience(j.nodeID, true) {
		return false
	}

	if !allow(claims) {
		return false
	}

	return true
}

func initJWT(durationHours int, nodeID string, privateKey []byte, userMap map[string]*users.UserData, claimValidator ClaimValidator) (*JWTAuth, func(context echo.Context) bool, MiddlewareValidator, error) {
	jwtAuth, err := NewJWTAuth(time.Duration(durationHours)*time.Hour, nodeID, privateKey)
	if err != nil {
		return nil, nil, nil, err
	}

	jwtAuthSkipper := func(context echo.Context) bool {
		path := context.Request().RequestURI

		if strings.HasSuffix(path, shared.AuthRoute()) || strings.HasSuffix(path, shared.AuthInfoRoute()) || path == "/" {
			return true
		}

		return false
	}

	jwtAuthAllow := func(e echo.Context, authContext *AuthContext) bool {
		isValidSubject := false

		for username := range userMap {
			if authContext.claims.VerifySubject(username) {
				isValidSubject = true
			}
		}

		if !claimValidator(authContext.claims) {
			return false
		}

		if !isValidSubject {
			return false
		}

		return true
	}

	return jwtAuth, jwtAuthSkipper, jwtAuthAllow, nil
}

func AddJWTAuth(webAPI WebAPI, config JWTAuthConfiguration, privateKey []byte, userMap map[string]*users.UserData, claimValidator ClaimValidator) *JWTAuth {
	durationHours := config.DurationHours

	// If durationHours is 0, we set 24h as the default durationHours.
	if durationHours <= 0 {
		durationHours = defaultJwtDurationHours
	}

	jwtAuth, jwtSkipper, jwtAuthAllow, _ := initJWT(durationHours, "wasp0", privateKey, userMap, claimValidator)

	webAPI.Use(jwtAuth.Middleware(jwtSkipper, jwtAuthAllow))

	return jwtAuth
}
