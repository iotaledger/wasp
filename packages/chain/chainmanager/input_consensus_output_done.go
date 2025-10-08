// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainmanager

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/chain/cmtlog"
	"github.com/iotaledger/wasp/v2/packages/chain/consensus"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
)

type inputConsensusOutputDone struct {
	committeeAddr   cryptolib.Address
	logIndex        cmtlog.LogIndex
	proposedBaseAO  *isc.StateAnchor
	consensusResult *consensus.Result
}

func NewInputConsensusOutputDone(
	committeeAddr cryptolib.Address,
	logIndex cmtlog.LogIndex,
	proposedBaseAO *isc.StateAnchor,
	consensusResult *consensus.Result,
) gpa.Input {
	return &inputConsensusOutputDone{
		committeeAddr:   committeeAddr,
		logIndex:        logIndex,
		proposedBaseAO:  proposedBaseAO,
		consensusResult: consensusResult,
	}
}

func (inp *inputConsensusOutputDone) String() string {
	return fmt.Sprintf(
		"{chainMgr.inputConsensusOutputDone, committeeAddr=%v, logIndex=%v, proposedBaseAO=%v, consensusResult=%v}",
		inp.committeeAddr.String(),
		inp.logIndex,
		inp.proposedBaseAO.Hash().Hex(),
		inp.consensusResult,
	)
}
