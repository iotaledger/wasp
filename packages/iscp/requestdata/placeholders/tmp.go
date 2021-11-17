// placehiolder for iota.go types
package placeholders

import (
	"github.com/iotaledger/hive.go/crypto/ed25519"
	iotago "github.com/iotaledger/iota.go/v3"
)

type UnknownOutput struct {
	Address iotago.Address
	Blocks  iotago.FeatureBlocks
}

// NewED25519Address creates a new ED25519Address from the given public key.
func NewED25519Address(publicKey ed25519.PublicKey) iotago.Address {
	// digest := blake2b.Sum256(publicKey[:])

	return nil
}
