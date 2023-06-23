package wallet

import (
	"fmt"

	"github.com/spf13/viper"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

var AddressIndex uint64

type Wallet struct {
	KeyPair      cryptolib.VariantKeyPair
	AddressIndex uint64
}

const (
	SchemeKeyChain = "keychain"
	SchemeLedger   = "ledger"
)

func GetWalletScheme() string {
	scheme := viper.GetString("wallet.scheme")

	fmt.Println(scheme)

	switch scheme {
	case SchemeLedger, SchemeKeyChain:
		return scheme
	default:
		return SchemeKeyChain
	}
}

func loadLedgerWallet() *Wallet {
	NewHWKeyPair()

	return &Wallet{}
}

func loadKeyChainWallet() *Wallet {
	keyChain := NewKeyChain()

	kp, err := keyChain.KeyPair(AddressIndex)
	log.Check(err)

	return &Wallet{KeyPair: kp, AddressIndex: AddressIndex}
}

func Load() *Wallet {
	walletScheme := GetWalletScheme()

	log.Printf("Scheme: %v", walletScheme)
	switch walletScheme {
	case SchemeKeyChain:
		return loadKeyChainWallet()
	case SchemeLedger:
		return loadLedgerWallet()
	}

	log.Fatal("invalid wallet scheme")
	return nil
	seedHex := viper.GetString("wallet.seed")

	useLegacyDerivation := viper.GetBool("wallet.useLegacyDerivation")
	if seedHex == "" {
		log.Fatal("call `init` first")
	}

	masterSeed, err := iotago.DecodeHex(seedHex)
	log.Check(err)

	kp := cryptolib.KeyPairFromSeed(cryptolib.SubSeed(masterSeed, uint32(AddressIndex), useLegacyDerivation))

	return &Wallet{KeyPair: kp, AddressIndex: AddressIndex}
}

func (w *Wallet) Address() iotago.Address {
	return w.KeyPair.Address()
}
