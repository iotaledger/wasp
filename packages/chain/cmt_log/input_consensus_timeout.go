// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmt_log

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
)

type inputConsensusTimeout struct {
	logIndex LogIndex
}

// NewInputConsensusTimeout creates an internal message that should be sent by other components (e.g. consensus or the chain).
func NewInputConsensusTimeout(logIndex LogIndex) gpa.Input {
	return &inputConsensusTimeout{
		logIndex: logIndex,
	}
}

func (inp *inputConsensusTimeout) String() string {
	return fmt.Sprintf(
		"{cmtLog.inputConsensusTimeout, logIndex=%v}",
		inp.logIndex,
	)
}
