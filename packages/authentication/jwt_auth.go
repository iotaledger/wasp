package authentication

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"

	"github.com/iotaledger/wasp/v2/packages/authentication/shared/permissions"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
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

func NewJWTAuth(duration time.Duration, nodeIDKeypair *cryptolib.KeyPair) *JWTAuth {
	return &JWTAuth{
		duration: duration,
		nodeID:   nodeIDKeypair.Address().String(),
		secret:   nodeIDKeypair.GetPrivateKey().AsBytes(),
	}
}

func (j *JWTAuth) IssueJWT(username string, claims *WaspClaims) (string, error) {
	now := time.Now()

	// Set claims
	registeredClaims := jwt.RegisteredClaims{
		Subject:   username,
		Issuer:    j.nodeID,
		Audience:  jwt.ClaimStrings{username},
		ID:        fmt.Sprintf("%d", now.Unix()),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
	}

	if j.duration > 0 {
		registeredClaims.ExpiresAt = jwt.NewNumericDate(now.Add(j.duration))
	}

	claims.RegisteredClaims = registeredClaims

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	return token.SignedString(j.secret)
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
