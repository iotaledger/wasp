// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainmanager

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/chain/cmt_log"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/sui-go/sui_types"
)

type inputConsensusOutputSkip struct {
	committeeAddr  cryptolib.Address
	logIndex       cmt_log.LogIndex
	proposedBaseAO sui_types.ObjectID
}

func NewInputConsensusOutputSkip(
	committeeAddr cryptolib.Address,
	logIndex cmt_log.LogIndex,
	proposedBaseAO sui_types.ObjectID,
) gpa.Input {
	return &inputConsensusOutputSkip{
		committeeAddr:  committeeAddr,
		logIndex:       logIndex,
		proposedBaseAO: proposedBaseAO,
	}
}

func (inp *inputConsensusOutputSkip) String() string {
	return fmt.Sprintf(
		"{chainMgr.inputConsensusOutputSkip, committeeAddr=%v, logIndex=%v, proposedBaseAO=%v}",
		inp.committeeAddr.String(),
		inp.logIndex,
		inp.proposedBaseAO.ToHex(),
	)
}
