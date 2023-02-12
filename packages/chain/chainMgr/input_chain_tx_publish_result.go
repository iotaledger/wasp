package chainMgr

import (
	"fmt"

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

func (i *inputChainTxPublishResult) String() string {
	return fmt.Sprintf(
		"{chainMgr.inputChainTxPublishResult, committeeAddr=%v, txID=%v, aliasOutput=%v, confirmed=%v}",
		i.committeeAddr.String(),
		i.txID.ToHex(),
		i.aliasOutput.OutputID().ToHex(),
		i.confirmed,
	)
}
