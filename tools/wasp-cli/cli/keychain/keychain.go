package keychain

import (
	"errors"
	"fmt"
	"time"

	"github.com/awnumar/memguard"

	"github.com/iotaledger/wasp/v2/packages/cryptolib"
)

const (
	strongholdKey     = "wasp-cli.stronghold.key"
	jwtTokenKeyPrefix = "wasp-cli.auth.jwt"
	seedKey           = "wasp-cli.seed"
)

const WaspCliServiceName = "IOTAFoundation.WaspCLI"

var (
	ErrKeyNotFound = errors.New("key not found")

	ErrTokenDoesNotExist      = errors.New("jwt token not found, call 'login'")
	ErrPasswordDoesNotExist   = errors.New("stronghold entry not found, call 'init'")
	ErrSeedDoesNotExist       = errors.New("seed not found, call 'init'")
	ErrSeedDoesNotMatchLength = errors.New("returned seed does not have a valid length")
)

type KeyChain interface {
	SetSeed(seed cryptolib.Seed) error
	GetSeed() (*cryptolib.Seed, error)

	SetStrongholdPassword(password *memguard.Enclave) error
	GetStrongholdPassword() (*memguard.Enclave, error)

	SetJWTAuthToken(node string, token string) error
	GetJWTAuthToken(node string) (string, error)
}

func jwtTokenKey(node string) string {
	return fmt.Sprintf("%s.%v", jwtTokenKeyPrefix, node)
}

func printWithTime(str string) {
	fmt.Printf("[%v] %v\n", time.Now(), str)
}
