package keychain

import (
	"encoding/json"
	"errors"
	"os"
	"path"
	"strings"

	"github.com/awnumar/memguard"
	jose "github.com/dvsekhvalnov/jose2go"

	"github.com/iotaledger/wasp/v2/packages/cryptolib"
)

const (
	fileName                   = "secrets.db"
	filePermission os.FileMode = 0o777
)

var ErrInvalidPassword = errors.New("invalid password")

type KeyChainFile struct {
	path             string
	passwordCallback func() *memguard.Enclave
	password         *memguard.Enclave
}

func NewKeyChainFile(path string, passwordCallback func() *memguard.Enclave) *KeyChainFile {
	return &KeyChainFile{
		path:             path,
		passwordCallback: passwordCallback,
	}
}

func (k *KeyChainFile) promptPassword() *memguard.Enclave {
	if k.password != nil {
		return k.password
	}

	enclave := k.passwordCallback()
	k.password = enclave

	return enclave
}

func (k *KeyChainFile) FilePath() string {
	return path.Join(k.path, fileName)
}

func (k *KeyChainFile) ReadContents() (map[string][]byte, error) {
	_, err := os.Stat(k.FilePath())

	if errors.Is(err, os.ErrNotExist) {
		result, neErr := json.Marshal(struct{}{})
		if neErr != nil {
			return nil, neErr
		}

		neErr = os.WriteFile(k.FilePath(), result, filePermission)
		if neErr != nil {
			return nil, neErr
		}

		return map[string][]byte{}, nil
	}

	content, err := os.ReadFile(k.FilePath())
	if err != nil {
		return nil, err
	}

	var secretsMap map[string][]byte
	err = json.Unmarshal(content, &secretsMap)
	if err != nil {
		return nil, err
	}

	return secretsMap, nil
}

func (k *KeyChainFile) Get(key string) ([]byte, error) {
	secrets, err := k.ReadContents()
	if err != nil {
		return nil, err
	}

	if val, ok := secrets[key]; ok {
		enclave := k.promptPassword()
		password, err := enclave.Open()
		if err != nil {
			return nil, err
		}
		defer password.Destroy()

		payload, _, err := jose.Decode(string(val), password.String())
		if err != nil {
			// There is no proper error definition available to check against, so it needs to be hardcoded here.
			if strings.Contains(err.Error(), "aes.KeyUnwrap") {
				return nil, ErrInvalidPassword
			}
			return nil, err
		}

		return []byte(payload), nil
	}

	return nil, ErrKeyNotFound
}

func (k *KeyChainFile) Set(key string, value []byte) error {
	secrets, err := k.ReadContents()
	if err != nil {
		return err
	}

	enclave := k.promptPassword()
	password, err := enclave.Open()
	if err != nil {
		return err
	}
	defer password.Destroy()

	token, err := jose.EncryptBytes(value, jose.PBES2_HS256_A128KW, jose.A256GCM, password.String())
	if err != nil {
		return err
	}

	secrets[key] = []byte(token)
	payload, err := json.MarshalIndent(secrets, "", " ")
	if err != nil {
		return err
	}

	err = os.WriteFile(k.FilePath(), payload, filePermission)
	if err != nil {
		return err
	}

	return nil
}

func (k *KeyChainFile) SetSeed(seed cryptolib.Seed) error {
	err := k.Set(seedKey, seed[:])
	return err
}

func (k *KeyChainFile) GetSeed() (*cryptolib.Seed, error) {
	seedItem, err := k.Get(seedKey)
	if err != nil {
		return nil, err
	}

	if len(seedItem) != cryptolib.SeedSize {
		return nil, ErrSeedDoesNotMatchLength
	}

	seed := cryptolib.SeedFromBytes(seedItem)
	return &seed, nil
}

func (k *KeyChainFile) SetStrongholdPassword(password *memguard.Enclave) error {
	buffer, err := password.Open()
	if err != nil {
		return err
	}
	defer buffer.Destroy()

	return k.Set(strongholdKey, buffer.Bytes())
}

func (k *KeyChainFile) GetStrongholdPassword() (*memguard.Enclave, error) {
	seedItem, err := k.Get(strongholdKey)
	if err != nil {
		return nil, err
	}

	return memguard.NewEnclave(seedItem), nil
}

func (k *KeyChainFile) SetJWTAuthToken(node string, token string) error {
	return k.Set(jwtTokenKey(node), []byte(token))
}

func (k *KeyChainFile) GetJWTAuthToken(node string) (string, error) {
	seedItem, err := k.Get(jwtTokenKey(node))
	// Special case. If the key is not found, return an empty token.
	if errors.Is(err, ErrKeyNotFound) {
		return "", nil
	} else if err != nil {
		return "", err
	}

	return string(seedItem), nil
}
