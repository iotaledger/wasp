package config

import (
	"errors"
	"fmt"

	"github.com/99designs/keyring"

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

type KeyChain struct {
	Keyring keyring.Keyring
}

func NewKeyChain() *KeyChain {
	ring, _ := keyring.Open(keyring.Config{
		ServiceName: "IOTAFoundation.WaspCLI",
	})

	return &KeyChain{
		Keyring: ring,
	}
}

func (k *KeyChain) SetSeed(seed cryptolib.Seed) error {
	err := k.Keyring.Set(keyring.Item{
		Key:  seedKey,
		Data: seed[:],
	})

	return err
}

func (k *KeyChain) GetSeed() (*cryptolib.Seed, error) {
	seedItem, err := k.Keyring.Get(seedKey)
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

func (k *KeyChain) SetStrongholdPassword(password string) error {
	return k.Keyring.Set(keyring.Item{
		Key:  strongholdKey,
		Data: []byte(password),
	})
}

func (k *KeyChain) GetStrongholdPassword() (string, error) {
	seedItem, err := k.Keyring.Get(strongholdKey)
	if errors.Is(err, keyring.ErrKeyNotFound) {
		return "", ErrPasswordDoesNotExist
	}

	return string(seedItem.Data), nil
}

func jwtTokenKey(node string) string {
	return fmt.Sprintf("%s.%s", jwtTokenKeyPrefix, node)
}

func (k *KeyChain) SetJWTAuthToken(node string, token string) error {
	return k.Keyring.Set(keyring.Item{
		Key:  jwtTokenKey(node),
		Data: []byte(token),
	})
}

func (k *KeyChain) GetJWTAuthToken(node string) (string, error) {
	seedItem, err := k.Keyring.Get(jwtTokenKey(node))
	// Special case. If the key is not found, return an empty token.
	if errors.Is(err, keyring.ErrKeyNotFound) {
		return "", nil
	}

	return string(seedItem.Data), nil
}
