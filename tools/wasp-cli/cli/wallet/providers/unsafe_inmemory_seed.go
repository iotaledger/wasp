package providers

import (
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet/wallets"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

type UnsafeInMemoryTestingSeed struct {
	cryptolib.VariantKeyPair
	addressIndex uint32
}

func newUnsafeInMemoryTestingSeed(keyPair *cryptolib.KeyPair, addressIndex uint32) *UnsafeInMemoryTestingSeed {
	return &UnsafeInMemoryTestingSeed{
		VariantKeyPair: keyPair,
		addressIndex:   addressIndex,
	}
}

func (i *UnsafeInMemoryTestingSeed) AddressIndex() uint32 {
	return i.addressIndex
}

func LoadUnsafeInMemoryTestingSeed(addressIndex uint32) wallets.Wallet {
	seed, err := hexutil.Decode(config.GetTestingSeed())
	log.Check(err)

	useLegacyDerivation := config.GetUseLegacyDerivation()
	keyPair := cryptolib.KeyPairFromSeed(cryptolib.SubSeed(seed, addressIndex, useLegacyDerivation))

	return newUnsafeInMemoryTestingSeed(keyPair, addressIndex)
}

func CreateUnsafeInMemoryTestingSeed() {
	seed := cryptolib.NewSeed()
	config.SetTestingSeed(hexutil.Encode(seed[:]))

	log.Printf("New testing seed saved inside the config file.\n")
}
