package keychain

import (
	"errors"
	"fmt"
	"runtime"

	"github.com/99designs/keyring"
	"github.com/awnumar/memguard"

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

type keyChain struct {
	Keyring keyring.Keyring
}

func newKeyRing() *keyChain {
	ring, _ := keyring.Open(keyring.Config{
		ServiceName: "IOTAFoundation.WaspCLI",
	})

	return &keyChain{
		Keyring: ring,
	}
}

func (k *keyChain) Close() {
	runtime.KeepAlive(k.Keyring)
	k.Keyring = nil
}

func SetSeed(seed cryptolib.Seed) error {
	store := newKeyRing()
	defer store.Close()

	err := store.Keyring.Set(keyring.Item{
		Key:  seedKey,
		Data: seed[:],
	})
	return err
}

func GetSeed() (*cryptolib.Seed, error) {
	store := newKeyRing()
	defer store.Close()

	seedItem, err := store.Keyring.Get(seedKey)
	if errors.Is(err, keyring.ErrKeyNotFound) {
		return nil, ErrSeedDoesNotExist
	}
	if err != nil {
		return nil, err
	}

	if len(seedItem.Data) != cryptolib.SeedSize {
		return nil, ErrSeedDoesNotMatchLength
	}

	var seed cryptolib.Seed
	copy(seed[:], seedItem.Data)

	return &seed, nil
}

func SetStrongholdPassword(password *memguard.Enclave) error {
	store := newKeyRing()
	defer store.Close()

	buffer, err := password.Open()
	if err != nil {
		return err
	}
	defer buffer.Destroy()

	return store.Keyring.Set(keyring.Item{
		Key:  strongholdKey,
		Data: buffer.Data(),
	})
}

func GetStrongholdPassword() (*memguard.Enclave, error) {
	store := newKeyRing()
	defer store.Close()

	seedItem, err := store.Keyring.Get(strongholdKey)
	if errors.Is(err, keyring.ErrKeyNotFound) {
		return nil, ErrPasswordDoesNotExist
	}

	return memguard.NewEnclave(seedItem.Data), nil
}

func jwtTokenKey(node string) string {
	return fmt.Sprintf("%s.%s", jwtTokenKeyPrefix, node)
}

func SetJWTAuthToken(node string, token string) error {
	store := newKeyRing()
	defer store.Close()

	return store.Keyring.Set(keyring.Item{
		Key:  jwtTokenKey(node),
		Data: []byte(token),
	})
}

func GetJWTAuthToken(node string) (string, error) {
	store := newKeyRing()
	defer store.Close()

	seedItem, err := store.Keyring.Get(jwtTokenKey(node))
	// Special case. If the key is not found, return an empty token.
	if errors.Is(err, keyring.ErrKeyNotFound) {
		return "", nil
	}

	return string(seedItem.Data), nil
}
