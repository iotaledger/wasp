// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtlog

import (
	"fmt"
)

type inputConsensusOutputSkip struct {
	logIndex LogIndex
}

// NewInputConsensusOutputSkip creates an internal message that should be sent by other components (e.g. consensus or the chain).
func NewInputConsensusOutputSkip(
	logIndex LogIndex,
) *inputConsensusOutputSkip {
	return &inputConsensusOutputSkip{
		logIndex: logIndex,
	}
}

func (inp *inputConsensusOutputSkip) String() string {
	return fmt.Sprintf(
		"{cmtLog.inputConsensusOutputSkip, logIndex=%v}",
		inp.logIndex,
	)
}
