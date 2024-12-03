// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmt_log

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/chain/cons"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type inputConsensusOutputDone struct {
	logIndex        LogIndex
	proposedBaseAO  *isc.StateAnchor // Proposed BaseAO
	consensusResult *cons.Result     // Decided BaseAO and the TX
}

// This message is internal one, but should be sent by other components (e.g. consensus or the chain).
func NewInputConsensusOutputDone(
	logIndex LogIndex,
	proposedBaseAO *isc.StateAnchor,
	consensusResult *cons.Result,
) gpa.Input {
	return &inputConsensusOutputDone{
		logIndex:        logIndex,
		proposedBaseAO:  proposedBaseAO,
		consensusResult: consensusResult,
	}
}

func (inp *inputConsensusOutputDone) String() string {
	return fmt.Sprintf(
		"{cmtLog.inputConsensusOutputDone, logIndex=%v, proposedBaseAO=%s, baseAnchorRef=%s}",
		inp.logIndex, inp.proposedBaseAO.Hash().Hex(), inp.consensusResult.DecidedAO.Hash().Hex(),
	)
}
