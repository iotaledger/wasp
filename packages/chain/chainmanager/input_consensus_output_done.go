// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainmanager

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/chain/cmt_log"
	"github.com/iotaledger/wasp/packages/chain/cons"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/sui-go/sui_types"
)

type inputConsensusOutputDone struct {
	committeeAddr   cryptolib.Address
	logIndex        cmt_log.LogIndex
	proposedBaseAO  sui_types.ObjectID
	consensusResult *cons.Result
}

func NewInputConsensusOutputDone(
	committeeAddr cryptolib.Address,
	logIndex cmt_log.LogIndex,
	proposedBaseAO sui_types.ObjectID,
	consensusResult *cons.Result,
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
		inp.proposedBaseAO.ToHex(),
		inp.consensusResult,
	)
}
