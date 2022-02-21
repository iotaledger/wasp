package jwt

import (
	"crypto/subtle"
	"fmt"
	jwt "github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"net/http"
	"time"
)

// Errors
var (
	ErrJWTInvalidClaims = echo.NewHTTPError(http.StatusUnauthorized, "invalid jwt claims")
	ErrInvalidJWT       = echo.NewHTTPError(http.StatusUnauthorized, "token is invalid")
)

type JWTAuth struct {
	sessionTimeout time.Duration
	nodeID         string
	secret         []byte
}

type ClaimValidator = func(claims *JWTAuthClaims) bool
type MiddlewareValidator = func(c echo.Context, claims *JWTAuthClaims) bool

// return error to continue to error routine, otherwise handle redirection or other stuff and return nil
type ErrorHandler = func(c echo.Context, error error) error

func NewJWTAuth(sessionTimeout time.Duration, nodeID string, secret []byte) (*JWTAuth, error) {
	return &JWTAuth{
		sessionTimeout: sessionTimeout,
		nodeID:         nodeID,
		secret:         secret,
	}, nil
}

type JWTAuthClaims struct {
	jwt.StandardClaims
	Dashboard  bool `json:"dashboard"`
	API        bool `json:"api"`
	ChainRead  bool `json:"chain.read"`
	ChainWrite bool `json:"chain.write"`
}

func (c *JWTAuthClaims) HasClaim(claim string) bool {
	var stdClaims jwt.Claims = c
	mapClaims := stdClaims.(jwt.MapClaims)

	if _, ok := mapClaims[claim]; ok {
		return true
	}

	return false
}

func (c *JWTAuthClaims) compare(field string, expected string) bool {
	if field == "" {
		return false
	}
	if subtle.ConstantTimeCompare([]byte(field), []byte(expected)) != 0 {
		return true
	}

	return false
}

func (c *JWTAuthClaims) VerifySubject(expected string) bool {
	return c.compare(c.Subject, expected)
}

func (j *JWTAuth) Middleware(skipper middleware.Skipper, allow MiddlewareValidator) echo.MiddlewareFunc {

	config := middleware.JWTConfig{
		ContextKey:  "jwt",
		Claims:      &JWTAuthClaims{},
		SigningKey:  j.secret,
		TokenLookup: "header:Authorization,cookie:jwt",
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {

		return func(c echo.Context) error {

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
			claims, ok := token.Claims.(*JWTAuthClaims)

			// do extended claims validation
			if !ok || !claims.VerifyAudience(j.nodeID, true) {
				return ErrJWTInvalidClaims
			}

			// validate claims
			if !allow(c, claims) {
				return ErrJWTInvalidClaims
			}

			// go to the next handler
			return next(c)
		}
	}
}

func (j *JWTAuth) IssueJWT(username string, authClaims *JWTAuthClaims) (string, error) {

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

	if j.sessionTimeout > 0 {
		stdClaims.ExpiresAt = now.Add(j.sessionTimeout).Unix()
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

	t, err := jwt.ParseWithClaims(token, &JWTAuthClaims{}, j.keyFunc)

	if err == nil && t.Valid {
		claims, ok := t.Claims.(*JWTAuthClaims)

		if !ok || !claims.VerifyAudience(j.nodeID, true) {
			return false
		}

		// validate claims
		if !allow(claims) {
			return false
		}

		return true
	}
	return false
}
