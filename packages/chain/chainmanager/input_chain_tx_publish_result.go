package chainmanager

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/packages/chain/cmtlog"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
)

type inputChainTxPublishResult struct {
	committeeAddr cryptolib.Address
	logIndex      cmtlog.LogIndex
	txDigest      iotago.Digest
	aliasOutput   *isc.StateAnchor
	confirmed     bool
}

func NewInputChainTxPublishResult(committeeAddr cryptolib.Address, logIndex cmtlog.LogIndex, txDigest iotago.Digest, aliasOutput *isc.StateAnchor, confirmed bool) gpa.Input {
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
