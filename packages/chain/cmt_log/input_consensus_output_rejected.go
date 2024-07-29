// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmt_log

import (
	"fmt"

	"github.com/iotaledger/wasp/clients/iscmove/iscmove_types"
	"github.com/iotaledger/wasp/packages/gpa"
)

type inputConsensusOutputRejected struct {
	aliasOutput *iscmove_types.Anchor
	logIndex    LogIndex
}

func NewInputConsensusOutputRejected(aliasOutput *iscmove_types.Anchor, logIndex LogIndex) gpa.Input {
	return &inputConsensusOutputRejected{
		aliasOutput: aliasOutput,
		logIndex:    logIndex,
	}
}

func (inp *inputConsensusOutputRejected) String() string {
	return fmt.Sprintf("{cmtLog.inputConsensusOutputRejected, %v, li=%v}", inp.aliasOutput, inp.logIndex)
}
