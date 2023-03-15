// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtLog

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type inputConsensusOutputConfirmed struct {
	aliasOutput *isc.AliasOutputWithID
	logIndex    LogIndex
}

func NewInputConsensusOutputConfirmed(aliasOutput *isc.AliasOutputWithID, logIndex LogIndex) gpa.Input {
	return &inputConsensusOutputConfirmed{
		aliasOutput: aliasOutput,
		logIndex:    logIndex,
	}
}

func (inp *inputConsensusOutputConfirmed) String() string {
	return fmt.Sprintf("{cmtLog.inputConsensusOutputConfirmed, %v, li=%v}", inp.aliasOutput, inp.logIndex)
}
