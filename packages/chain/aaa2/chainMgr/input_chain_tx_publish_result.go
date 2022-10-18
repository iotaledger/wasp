package chainMgr

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/gpa"
)

type inputChainTxPublishResult struct {
	committeeID CommitteeID
	txID        iotago.TransactionID
	confirmed   bool
}

func NewInputChainTxPublishResult(committeeID CommitteeID, txID iotago.TransactionID, confirmed bool) gpa.Input {
	return &inputChainTxPublishResult{
		committeeID: committeeID,
		txID:        txID,
		confirmed:   confirmed,
	}
}
