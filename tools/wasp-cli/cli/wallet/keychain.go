package wallet

import (
	"errors"
	"math"
	"math/rand"
	"time"

	"github.com/99designs/keyring"

	"github.com/iotaledger/wasp/packages/cryptolib"
)

const (
	strongholdKey = "wasp-cli.stronghold.key"
	jwtTokenKey   = "wasp-cli.auth.jwt"
	seedKey       = "wasp-cli.seed"
)

func randomizeBuffer(data *[]byte) {
	// TODO: Validate seed source
	randSeed := rand.NewSource(time.Now().UnixNano())
	random := rand.New(randSeed)

	for i := 0; i < len(*data); i++ {
		(*data)[i] = byte(random.Intn(math.MaxUint8 + 1))
	}
}

func randomizeSeed(data *cryptolib.Seed) {
	// TODO: Validate seed source
	randSeed := rand.NewSource(time.Now().UnixNano())
	random := rand.New(randSeed)

	for i := 0; i < len(*data); i++ {
		(*data)[i] = byte(random.Intn(math.MaxUint8 + 1))
	}
}

var (
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

func (k *KeyChain) InitializeKeyPair() error {
	seed := cryptolib.NewSeed()
	err := k.Keyring.Set(keyring.Item{
		Key:  seedKey,
		Data: seed[:],
	})

	return err
}

func (k *KeyChain) KeyPair(addressIndex uint64) (*cryptolib.KeyPair, error) {
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

	randomizeBuffer(&seedItem.Data)
	keyPair := cryptolib.KeyPairFromSeed(seed.SubSeed(addressIndex))

	return keyPair, nil
}
