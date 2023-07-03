package providers

import (
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet/wallets"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

type InMemoryWallet struct {
	cryptolib.VariantKeyPair
	addressIndex uint32
}

func newInMemoryWallet(keyPair *cryptolib.KeyPair, addressIndex uint32) *InMemoryWallet {
	return &InMemoryWallet{
		VariantKeyPair: keyPair,
		addressIndex:   addressIndex,
	}
}

func (i *InMemoryWallet) AddressIndex() uint32 {
	return i.addressIndex
}

func LoadInMemory(addressIndex uint32) wallets.Wallet {
	keyChain := config.NewKeyChain()
	seed, err := keyChain.GetSeed()
	log.Check(err)

	keyPair := cryptolib.KeyPairFromSeed(seed.SubSeed(uint64(addressIndex)))

	return newInMemoryWallet(keyPair, addressIndex)
}

func CreateNewInMemory() {
	keyChain := config.NewKeyChain()
	seed := cryptolib.NewSeed()
	err := keyChain.SetSeed(seed)
	log.Check(err)

	log.Printf("In memory seed saved in the keychain.\n")
}
