package wallets

import "github.com/iotaledger/wasp/packages/cryptolib"

type Wallet interface {
	cryptolib.VariantKeyPair

	AddressIndex() uint32
}
