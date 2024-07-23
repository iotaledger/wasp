package chainmanager

import (
	"fmt"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/cmt_log"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc/sui"
)

type inputChainTxPublishResult struct {
	committeeAddr cryptolib.Address
	logIndex      cmt_log.LogIndex
	txID          iotago.TransactionID
	aliasOutput   *sui.Anchor
	confirmed     bool
}

func NewInputChainTxPublishResult(committeeAddr cryptolib.Address, logIndex cmt_log.LogIndex, txID iotago.TransactionID, aliasOutput *sui.Anchor, confirmed bool) gpa.Input {
	return &inputChainTxPublishResult{
		committeeAddr: committeeAddr,
		logIndex:      logIndex,
		txID:          txID,
		aliasOutput:   aliasOutput,
		confirmed:     confirmed,
	}
}

func (i *inputChainTxPublishResult) String() string {
	return fmt.Sprintf(
		"{chainMgr.inputChainTxPublishResult, committeeAddr=%v, logIndex=%v, txID=%v, aliasOutput=%v, confirmed=%v}",
		i.committeeAddr.String(),
		i.logIndex,
		i.txID.ToHex(),
		i.aliasOutput,
		i.confirmed,
	)
}
