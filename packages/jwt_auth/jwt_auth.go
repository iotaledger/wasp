package jwt_auth

import (
	"crypto/subtle"
	"fmt"
	jwt "github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/libp2p/go-libp2p-core/crypto"
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

type ClaimValidator = func(claims *AuthClaims) bool
type MiddlewareValidator = func(c echo.Context, claims *AuthClaims) bool

// return error to continue to error routine, otherwise handle redirection or other stuff and return nil
type ErrorHandler = func(c echo.Context, error error) error

func NewJWTAuth(sessionTimeout time.Duration, nodeID string, secret crypto.PrivKey) (*JWTAuth, error) {
	secretBytes, err := crypto.MarshalPrivateKey(secret)
	if err != nil {
		return nil, fmt.Errorf("unable to convert private key: %w", err)
	}

	return &JWTAuth{
		sessionTimeout: sessionTimeout,
		nodeID:         nodeID,
		secret:         secretBytes,
	}, nil
}

type AuthClaims struct {
	jwt.StandardClaims
	Dashboard bool `json:"dashboard"`
	API       bool `json:"api"`
}

func (c *AuthClaims) HasClaim(claim string) bool {
	var stdClaims jwt.Claims = c
	mapClaims := stdClaims.(jwt.MapClaims)

	if _, ok := mapClaims[claim]; ok {
		return true
	}

	return false
}

func (c *AuthClaims) compare(field string, expected string) bool {
	if field == "" {
		return false
	}
	if subtle.ConstantTimeCompare([]byte(field), []byte(expected)) != 0 {
		return true
	}

	return false
}

func (c *AuthClaims) VerifySubject(expected string) bool {
	return c.compare(c.Subject, expected)
}

type AuthenticationMethod int

const (
	AuthenticateWebAPI AuthenticationMethod = iota
	AuthenticateDashboard
)

func (j *JWTAuth) Middleware(skipper middleware.Skipper, allow MiddlewareValidator) echo.MiddlewareFunc {

	config := middleware.JWTConfig{
		ContextKey: "jwt",
		Claims:     &AuthClaims{},
		SigningKey: j.secret,
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

			preToken := c.Get("jwt")
			token := preToken.(*jwt.Token)

			// validate the signing method we expect
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return ErrInvalidJWT
			}

			// read the claims set by the JWT middleware on the context
			claims, ok := token.Claims.(*AuthClaims)

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

func (j *JWTAuth) IssueJWT(username string, api bool, dashboard bool) (string, error) {

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

	claims := &AuthClaims{
		StandardClaims: stdClaims,
		Dashboard:      dashboard,
		API:            api,
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	return token.SignedString(j.secret)
}

func (j *JWTAuth) VerifyJWT(token string, allow ClaimValidator) bool {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		// validate the signing method we expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return j.secret, nil
	}

	t, err := jwt.ParseWithClaims(token, &AuthClaims{}, keyFunc)

	if err == nil && t.Valid {
		claims, ok := t.Claims.(*AuthClaims)

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
