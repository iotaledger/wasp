package providers

import (
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet/wallets"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

type KeyChainWallet struct {
	cryptolib.VariantKeyPair
	addressIndex uint32
}

func newInMemoryWallet(keyPair *cryptolib.KeyPair, addressIndex uint32) *KeyChainWallet {
	return &KeyChainWallet{
		VariantKeyPair: keyPair,
		addressIndex:   addressIndex,
	}
}

func (i *KeyChainWallet) AddressIndex() uint32 {
	return i.addressIndex
}

func LoadKeyChain(addressIndex uint32) wallets.Wallet {
	seed, err := config.GetKeyChain().GetSeed()
	log.Check(err)

	useLegacyDerivation := config.GetUseLegacyDerivation()
	keyPair := cryptolib.KeyPairFromSeed(cryptolib.SubSeed(seed[:], addressIndex, useLegacyDerivation))

	return newInMemoryWallet(keyPair, addressIndex)
}

func CreateKeyChain() {
	seed := cryptolib.NewSeed()
	err := config.GetKeyChain().SetSeed(seed)
	log.Check(err)

	log.Printf("New seed stored in the keychain.\n")
}

func MigrateKeyChain(seed cryptolib.Seed) {
	err := config.GetKeyChain().SetSeed(seed)
	log.Check(err)
	log.Printf("Seed migrated to keychain")
}
