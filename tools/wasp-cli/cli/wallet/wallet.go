package wallet

import (
	"github.com/spf13/viper"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

var AddressIndex int

type Wallet struct {
	KeyPair      *cryptolib.KeyPair
	AddressIndex int
}

func Load() *Wallet {
	seedHex := viper.GetString("wallet.seed")

	if seedHex == "" {
		log.Fatal("call `init` first")
	}

	seedBytes, err := iotago.DecodeHex(seedHex)
	log.Check(err)

	seed := cryptolib.NewSeedFromBytes(seedBytes)
	kp := cryptolib.NewKeyPairFromSeed(seed.SubSeed(uint64(AddressIndex)))

	return &Wallet{KeyPair: kp, AddressIndex: AddressIndex}
}

func (w *Wallet) PrivateKey() *cryptolib.PrivateKey {
	return w.KeyPair.GetPrivateKey()
}

func (w *Wallet) Address() iotago.Address {
	return w.KeyPair.GetPublicKey().AsEd25519Address()
}
