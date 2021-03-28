// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"time"
)

const (
	initialTimeoutPullInclusionState = 2 * time.Second
	periodPullInclusionStage         = 5 * time.Second
)

var nilTxID ledgerstate.TransactionID

func (op *operator) setFinalizedTransaction(txid ledgerstate.TransactionID) {
	if op.postedResultTxid != nilTxID {
		op.log.Warn("duplicated transaction")
	}
	op.postedResultTxid = txid
	op.nextPullInclusionLevel = time.Now().Add(initialTimeoutPullInclusionState)
	op.log.Debugf("finalized tx set to %s", txid.String())
}

func (op *operator) checkInclusionLevel(txid *ledgerstate.TransactionID, level ledgerstate.InclusionState) {
	if op.postedResultTxid == nilTxID {
		return
	}
	if op.postedResultTxid != *txid {
		return
	}
	switch level {
	case ledgerstate.Pending:
		if op.consensusStage != consensusStageResultTransactionBooked {
			op.setNextConsensusStage(consensusStageResultTransactionBooked)
		} else {
			op.setNextPullInclusionStageDeadline()
		}
	case ledgerstate.Rejected:
		// cannot move to the next leader because funds are locked forever
		// TODO not clear what to do. Need proper specs from Goshimmer
		op.log.Warnf("!!!!!!! received 'rejected' for transaction %s. Not clear what to do. Need proper specs from Goshimmer",
			op.postedResultTxid.String())
	}
}

func (op *operator) setNextPullInclusionStageDeadline() {
	op.nextPullInclusionLevel = time.Now().Add(periodPullInclusionStage)
}
