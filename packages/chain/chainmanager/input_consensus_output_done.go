// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainmanager

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/chain/committeelog"
	"github.com/iotaledger/wasp/v2/packages/chain/consensus"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
)

type inputConsensusOutputDone struct {
	committeeAddr      cryptolib.Address
	logIndex           committeelog.LogIndex
	proposedBaseAnchor *isc.StateAnchor
	consensusResult    *consensus.Result
}

func NewInputConsensusOutputDone(
	committeeAddr cryptolib.Address,
	logIndex committeelog.LogIndex,
	proposedBaseAnchor *isc.StateAnchor,
	consensusResult *consensus.Result,
) gpa.Input {
	return &inputConsensusOutputDone{
		committeeAddr:      committeeAddr,
		logIndex:           logIndex,
		proposedBaseAnchor: proposedBaseAnchor,
		consensusResult:    consensusResult,
	}
}

func (inp *inputConsensusOutputDone) String() string {
	return fmt.Sprintf(
		"{chainMgr.inputConsensusOutputDone, committeeAddr=%v, logIndex=%v, proposedBaseAnchor=%v, consensusResult=%v}",
		inp.committeeAddr.String(),
		inp.logIndex,
		inp.proposedBaseAnchor.Hash().Hex(),
		inp.consensusResult,
	)
}
