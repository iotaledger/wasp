package keychain

import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/awnumar/memguard"
	"github.com/zalando/go-keyring"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

const (
	strongholdKey     = "wasp-cli.stronghold.key"
	jwtTokenKeyPrefix = "wasp-cli.auth.jwt"
	seedKey           = "wasp-cli.seed"
)

var (
	ErrTokenDoesNotExist      = errors.New("jwt token not found, call 'login'")
	ErrPasswordDoesNotExist   = errors.New("stronghold entry not found, call 'init'")
	ErrSeedDoesNotExist       = errors.New("seed not found, call 'init'")
	ErrSeedDoesNotMatchLength = errors.New("returned seed does not have a valid length")
)

const WaspCliServiceName = "IOTAFoundation.WaspCLI"

func SetSeed(seed cryptolib.Seed) error {
	err := keyring.Set(WaspCliServiceName, seedKey, iotago.EncodeHex(seed[:]))
	return err
}

func GetSeed() (*cryptolib.Seed, error) {
	seedItem, err := keyring.Get(WaspCliServiceName, seedKey)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil, ErrSeedDoesNotExist
	}
	if err != nil {
		return nil, err
	}

	seedBytes, err := iotago.DecodeHex(seedItem)
	if len(seedBytes) != cryptolib.SeedSize {
		return nil, ErrSeedDoesNotMatchLength
	}

	seed := cryptolib.SeedFromBytes(seedBytes)
	return &seed, nil
}

func SetStrongholdPassword(password *memguard.Enclave) error {
	buffer, err := password.Open()
	if err != nil {
		return err
	}
	defer buffer.Destroy()

	return keyring.Set(WaspCliServiceName, strongholdKey, buffer.String())
}

func GetStrongholdPassword() (*memguard.Enclave, error) {
	enclave, err := Get("service", "user")
	if err != nil {
		panic(err)
	}

	secretPassword, err := enclave.Open()
	if err != nil {
		panic(err)
	}
	defer secretPassword.Destroy()

	seedItem, err := keyring.Get(WaspCliServiceName, strongholdKey)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil, ErrPasswordDoesNotExist
	}

	memguard.WipeBytes(*(*[]byte)(unsafe.Pointer(&seedItem)))

	return memguard.NewEnclave([]byte(seedItem)), nil
}

func jwtTokenKey(node string) string {
	return fmt.Sprintf("%s.%s", jwtTokenKeyPrefix, node)
}

func SetJWTAuthToken(node string, token string) error {
	return keyring.Set(WaspCliServiceName, jwtTokenKey(node), token)
}

func GetJWTAuthToken(node string) (string, error) {
	seedItem, err := keyring.Get(WaspCliServiceName, jwtTokenKey(node))
	// Special case. If the key is not found, return an empty token.
	if errors.Is(err, keyring.ErrNotFound) {
		return "", nil
	}

	return seedItem, nil
}
