// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtLog

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type inputConsensusOutputRejected struct {
	aliasOutput *isc.AliasOutputWithID
	logIndex    LogIndex
}

func NewInputConsensusOutputRejected(aliasOutput *isc.AliasOutputWithID, logIndex LogIndex) gpa.Input {
	return &inputConsensusOutputRejected{
		aliasOutput: aliasOutput,
		logIndex:    logIndex,
	}
}

func (inp *inputConsensusOutputRejected) String() string {
	return fmt.Sprintf("{cmtLog.inputConsensusOutputRejected, %v, li=%v}", inp.aliasOutput, inp.logIndex)
}
