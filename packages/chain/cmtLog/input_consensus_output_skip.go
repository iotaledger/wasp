// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtLog

import (
	"fmt"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/gpa"
)

type inputConsensusOutputSkip struct {
	logIndex       LogIndex
	proposedBaseAO iotago.OutputID
}

// This message is internal one, but should be sent by other components (e.g. consensus or the chain).
func NewInputConsensusOutputSkip(
	logIndex LogIndex,
	proposedBaseAO iotago.OutputID,
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
