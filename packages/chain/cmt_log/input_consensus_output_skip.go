// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmt_log

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/sui-go/sui"
)

type inputConsensusOutputSkip struct {
	logIndex       LogIndex
	proposedBaseAO *sui.ObjectRef
}

// This message is internal one, but should be sent by other components (e.g. consensus or the chain).
func NewInputConsensusOutputSkip(
	logIndex LogIndex,
	proposedBaseAO *sui.ObjectRef,
) gpa.Input {
	return &inputConsensusOutputSkip{
		logIndex:       logIndex,
		proposedBaseAO: proposedBaseAO,
	}
}

func (inp *inputConsensusOutputSkip) String() string {
	return fmt.Sprintf(
		"{cmtLog.inputConsensusOutputSkip, logIndex=%v, proposedBaseAO=%s}",
		inp.logIndex, inp.proposedBaseAO.String(),
	)
}
