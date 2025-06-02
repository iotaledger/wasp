// Package wallets provides wallet management functionality for wasp-cli,
// allowing users to interact with cryptocurrency wallets.
package wallets

import "github.com/iotaledger/wasp/packages/cryptolib"

type Wallet interface {
	cryptolib.Signer

	AddressIndex() uint32
}
