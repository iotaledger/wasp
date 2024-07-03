package transaction

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
)

// TODO: Keeping it to give context for further refactoring
type MintNFTsTransactionParams struct {
	IssuerKeyPair      cryptolib.Signer
	CollectionOutputID *iotago.OutputID
	Target             *cryptolib.Address
	ImmutableMetadata  [][]byte
	UnspentOutputs     iotago.OutputSet
	UnspentOutputIDs   iotago.OutputIDs
}
