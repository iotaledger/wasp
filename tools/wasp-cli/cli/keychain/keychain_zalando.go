// Package keychain provides secure management of cryptographic keys
// and credentials for the wasp-cli tool.
package keychain

import (
	"errors"

	"github.com/awnumar/memguard"
	"github.com/zalando/go-keyring"

	"github.com/iotaledger/wasp/packages/cryptolib"
)

type KeyChainZalando struct{}

func NewKeyChainZalando() *KeyChainZalando {
	return &KeyChainZalando{}
}

// IsKeyChainAvailable validates existence of a keychain by querying an entry.
// If a keychain is not available, it will throw an internal OS error,
// while an existing keychain will return ErrNotFound (as the key does not exist)
func IsKeyChainAvailable() bool {
	_, err := keyring.Get("", "")

	if errors.Is(err, keyring.ErrNotFound) {
		return true
	}

	if errors.Is(err, keyring.ErrUnsupportedPlatform) {
		return false
	}

	return false
}

func (k *KeyChainZalando) SetSeed(seed cryptolib.Seed) error {
	err := keyring.Set(WaspCliServiceName, seedKey, cryptolib.EncodeHex(seed[:]))
	return err
}

func (k *KeyChainZalando) GetSeed() (*cryptolib.Seed, error) {
	seedItem, err := keyring.Get(WaspCliServiceName, seedKey)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil, ErrSeedDoesNotExist
	}
	if err != nil {
		return nil, err
	}

	seedBytes, err := cryptolib.DecodeHex(seedItem)
	if err != nil {
		return nil, err
	}

	if len(seedBytes) != cryptolib.SeedSize {
		return nil, ErrSeedDoesNotMatchLength
	}

	seed := cryptolib.SeedFromBytes(seedBytes)
	return &seed, nil
}

func (k *KeyChainZalando) SetStrongholdPassword(password *memguard.Enclave) error {
	buffer, err := password.Open()
	if err != nil {
		return err
	}
	defer buffer.Destroy()

	return keyring.Set(WaspCliServiceName, strongholdKey, buffer.String())
}

func (k *KeyChainZalando) GetStrongholdPassword() (*memguard.Enclave, error) {
	seedItem, err := keyring.Get(WaspCliServiceName, strongholdKey)
	if errors.Is(err, keyring.ErrNotFound) {
		return nil, ErrPasswordDoesNotExist
	}

	return memguard.NewEnclave([]byte(seedItem)), nil
}

func (k *KeyChainZalando) SetJWTAuthToken(node string, token string) error {
	return keyring.Set(WaspCliServiceName, jwtTokenKey(node), token)
}

func (k *KeyChainZalando) GetJWTAuthToken(node string) (string, error) {
	seedItem, err := keyring.Get(WaspCliServiceName, jwtTokenKey(node))
	// Special case. If the key is not found, return an empty token.
	if errors.Is(err, keyring.ErrNotFound) {
		return "", nil
	}

	return seedItem, nil
}
