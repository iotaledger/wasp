// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmt_log

import (
	"fmt"

	"github.com/iotaledger/wasp/clients/iscmove/isctypes"
	"github.com/iotaledger/wasp/packages/gpa"
)

type inputConsensusOutputConfirmed struct {
	aliasOutput *isctypes.Anchor
	logIndex    LogIndex
}

func NewInputConsensusOutputConfirmed(aliasOutput *isctypes.Anchor, logIndex LogIndex) gpa.Input {
	return &inputConsensusOutputConfirmed{
		aliasOutput: aliasOutput,
		logIndex:    logIndex,
	}
}

func (inp *inputConsensusOutputConfirmed) String() string {
	return fmt.Sprintf("{cmtLog.inputConsensusOutputConfirmed, %v, li=%v}", inp.aliasOutput, inp.logIndex)
}
