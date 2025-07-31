// Package providers implements different wallet provider implementations
// that can be used by the wasp-cli wallet commands.
package providers

import (
	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/wallet/wallets"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
)

type UnsafeInMemoryTestingSeed struct {
	cryptolib.Signer
	addressIndex uint32
}

func NewUnsafeInMemoryTestingSeed(keyPair *cryptolib.KeyPair, addressIndex uint32) *UnsafeInMemoryTestingSeed {
	return &UnsafeInMemoryTestingSeed{
		Signer:       keyPair,
		addressIndex: addressIndex,
	}
}

func (i *UnsafeInMemoryTestingSeed) AddressIndex() uint32 {
	return i.addressIndex
}

func LoadUnsafeInMemoryTestingSeed(addressIndex uint32) wallets.Wallet {
	seed, err := hexutil.Decode(config.GetTestingSeed())
	log.Check(err)

	keyPair := cryptolib.KeyPairFromSeed(cryptolib.SubSeed(seed, addressIndex))

	return NewUnsafeInMemoryTestingSeed(keyPair, addressIndex)
}

func CreateUnsafeInMemoryTestingSeed() {
	seed := cryptolib.NewSeed()
	config.SetTestingSeed(hexutil.Encode(seed[:]))

	log.Printf("New testing seed saved inside the config file.\n")
}
