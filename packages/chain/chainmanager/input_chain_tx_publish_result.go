package chainmanager

import (
	"fmt"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/chain/cmt_log"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type inputChainTxPublishResult struct {
	committeeAddr cryptolib.Address
	logIndex      cmt_log.LogIndex
	txDigest      iotago.Digest
	aliasOutput   *isc.StateAnchor
	confirmed     bool
}

func NewInputChainTxPublishResult(committeeAddr cryptolib.Address, logIndex cmt_log.LogIndex, txDigest iotago.Digest, aliasOutput *isc.StateAnchor, confirmed bool) gpa.Input {
	return &inputChainTxPublishResult{
		committeeAddr: committeeAddr,
		logIndex:      logIndex,
		txDigest:      txDigest,
		aliasOutput:   aliasOutput,
		confirmed:     confirmed,
	}
}

func (i *inputChainTxPublishResult) String() string {
	return fmt.Sprintf(
		"{chainMgr.inputChainTxPublishResult, committeeAddr=%v, logIndex=%v, txDigest=%s, aliasOutput=%v, confirmed=%v}",
		i.committeeAddr.String(),
		i.logIndex,
		i.txDigest,
		i.aliasOutput,
		i.confirmed,
	)
}
