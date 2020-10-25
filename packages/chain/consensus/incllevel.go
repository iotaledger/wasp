package consensus

import (
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/goshimmer/dapps/waspconn/packages/waspconn"
	"time"
)

const (
	initialTimeoutPullInclusionState = 2 * time.Second
	periodPullInclusionStage         = 5 * time.Second
)

func (op *operator) setFinalizedTransaction(txid *valuetransaction.ID) {
	if op.postedResultTxid != nil {
		op.log.Warn("duplicated transaction to follow")
	}
	op.postedResultTxid = txid
	op.nextPullInclusionLevel = time.Now().Add(initialTimeoutPullInclusionState)
	op.log.Debugf("finalized tx set to %s", txid.String())
}

func (op *operator) checkInclusionLevel(txid *valuetransaction.ID, level byte) {
	if op.postedResultTxid == nil {
		return
	}
	if *op.postedResultTxid != *txid {
		return
	}
	switch level {
	case waspconn.TransactionInclusionLevelBooked:
		if op.consensusStage != consensusStageResultTransactionBooked {
			op.setNextConsensusStage(consensusStageResultTransactionBooked)
		} else {
			op.setNextPullInclusionStageDeadline()
		}
	case waspconn.TransactionInclusionLevelRejected:
		// cannot move to the next leader because funds are locked forever
		// TODO not clear what to do. Need proper specs from Goshimmer
		op.log.Warnf("!!!!!!! received 'rejected' for transaction %s. Not clear what to do. Need proper specs from Goshimmer",
			op.postedResultTxid.String())
	}
}

func (op *operator) setNextPullInclusionStageDeadline() {
	op.nextPullInclusionLevel = time.Now().Add(periodPullInclusionStage)
}
