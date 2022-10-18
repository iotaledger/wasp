package chainMgr

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type inputChainTxPublishResult struct {
	committeeID CommitteeID
	txID        iotago.TransactionID
	aliasOutput *isc.AliasOutputWithID
	confirmed   bool
}

func NewInputChainTxPublishResult(committeeID CommitteeID, txID iotago.TransactionID, aliasOutput *isc.AliasOutputWithID, confirmed bool) gpa.Input {
	return &inputChainTxPublishResult{
		committeeID: committeeID,
		txID:        txID,
		aliasOutput: aliasOutput,
		confirmed:   confirmed,
	}
}
