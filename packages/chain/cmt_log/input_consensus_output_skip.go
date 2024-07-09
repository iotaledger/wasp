// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmt_log

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/sui-go/sui_types"
)

type inputConsensusOutputSkip struct {
	logIndex       LogIndex
	proposedBaseAO sui_types.ObjectID
}

// This message is internal one, but should be sent by other components (e.g. consensus or the chain).
func NewInputConsensusOutputSkip(
	logIndex LogIndex,
	proposedBaseAO sui_types.ObjectID,
) gpa.Input {
	return &inputConsensusOutputSkip{
		logIndex:       logIndex,
		proposedBaseAO: proposedBaseAO,
	}
}

func (inp *inputConsensusOutputSkip) String() string {
	return fmt.Sprintf(
		"{cmtLog.inputConsensusOutputSkip, logIndex=%v, proposedBaseAO=%v}",
		inp.logIndex, inp.proposedBaseAO.ToHex(),
	)
}
