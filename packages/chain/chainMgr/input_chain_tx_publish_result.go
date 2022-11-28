package chainMgr

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type inputChainTxPublishResult struct {
	committeeAddr iotago.Ed25519Address
	txID          iotago.TransactionID
	aliasOutput   *isc.AliasOutputWithID
	confirmed     bool
}

func NewInputChainTxPublishResult(committeeAddr iotago.Ed25519Address, txID iotago.TransactionID, aliasOutput *isc.AliasOutputWithID, confirmed bool) gpa.Input {
	return &inputChainTxPublishResult{
		committeeAddr: committeeAddr,
		txID:          txID,
		aliasOutput:   aliasOutput,
		confirmed:     confirmed,
	}
}
