package keychain

import (
	"errors"
	"runtime"
	"syscall"

	"github.com/99designs/keyring"
	"github.com/awnumar/memguard"
	"golang.org/x/term"

	"github.com/iotaledger/wasp/v2/packages/cryptolib"
)

/**
Will most likely be removed and replaced by zalando/go-keyring + `keychain_file.go`
*/

type KeyChain99 struct {
	Keyring keyring.Keyring
}

func passwordCallback(m string) (string, error) {
	printWithTime("Enter password to unlock the keychain")
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin)) //nolint:unconvert // int cast is needed for windows
	printWithTime("")

	return string(passwordBytes), err
}

func NewKeyRing99(baseDir string) *KeyChain99 {
	printWithTime("newKeyRing")

	printWithTime("availableBackends:")
	for _, b := range keyring.AvailableBackends() {
		printWithTime(string(b))
	}

	ring, _ := keyring.Open(keyring.Config{
		ServiceName:          WaspCliServiceName,
		FileDir:              baseDir,
		FilePasswordFunc:     passwordCallback,
		KeychainPasswordFunc: passwordCallback,
	})
	printWithTime("newKeyRing opened")

	return &KeyChain99{
		Keyring: ring,
	}
}

func (k *KeyChain99) Close() {
	runtime.KeepAlive(k.Keyring)
	k.Keyring = nil
}

func (k *KeyChain99) SetSeed(seed cryptolib.Seed) error {
	printWithTime("SetSeed start")

	err := k.Keyring.Set(keyring.Item{
		Key:  seedKey,
		Data: seed[:],
	})

	printWithTime("SetSeed finished")

	return err
}

func (k *KeyChain99) GetSeed() (*cryptolib.Seed, error) {
	printWithTime("GetSeed start")

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
	printWithTime("GetSeed finished")

	return &seed, nil
}

func (k *KeyChain99) SetStrongholdPassword(password *memguard.Enclave) error {
	buffer, err := password.Open()
	if err != nil {
		return err
	}
	defer buffer.Destroy()

	return k.Keyring.Set(keyring.Item{
		Key:  strongholdKey,
		Data: buffer.Data(),
	})
}

func (k *KeyChain99) GetStrongholdPassword() (*memguard.Enclave, error) {
	seedItem, err := k.Keyring.Get(strongholdKey)
	if errors.Is(err, keyring.ErrKeyNotFound) {
		return nil, ErrPasswordDoesNotExist
	}

	return memguard.NewEnclave(seedItem.Data), nil
}

func (k *KeyChain99) SetJWTAuthToken(node string, token string) error {
	printWithTime("SetJWTAuthToken start")

	err := k.Keyring.Set(keyring.Item{
		Key:  jwtTokenKey(node),
		Data: []byte(token),
	})

	printWithTime("SetJWTAuthToken finished")

	return err
}

func (k *KeyChain99) GetJWTAuthToken(node string) (string, error) {
	printWithTime("GetJWTAuthToken start")

	seedItem, err := k.Keyring.Get(jwtTokenKey(node))
	// Special case. If the key is not found, return an empty token.
	if errors.Is(err, keyring.ErrKeyNotFound) {
		return "", nil
	}
	printWithTime("GetJWTAuthToken finished")

	return string(seedItem.Data), nil
}
