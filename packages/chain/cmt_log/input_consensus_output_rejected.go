// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmt_log

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc/sui"
)

type inputConsensusOutputRejected struct {
	aliasOutput *sui.Anchor
	logIndex    LogIndex
}

func NewInputConsensusOutputRejected(aliasOutput *sui.Anchor, logIndex LogIndex) gpa.Input {
	return &inputConsensusOutputRejected{
		aliasOutput: aliasOutput,
		logIndex:    logIndex,
	}
}

func (inp *inputConsensusOutputRejected) String() string {
	return fmt.Sprintf("{cmtLog.inputConsensusOutputRejected, %v, li=%v}", inp.aliasOutput, inp.logIndex)
}
