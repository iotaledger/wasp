// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package committeelog

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
)

type inputConsensusOutputRejected struct {
	aliasOutput *isc.StateAnchor
	logIndex    LogIndex
}

func NewInputConsensusOutputRejected(aliasOutput *isc.StateAnchor, logIndex LogIndex) gpa.Input {
	return &inputConsensusOutputRejected{
		aliasOutput: aliasOutput,
		logIndex:    logIndex,
	}
}

func (inp *inputConsensusOutputRejected) String() string {
	return fmt.Sprintf("{committeeLog.inputConsensusOutputRejected, %v, li=%v}", inp.aliasOutput, inp.logIndex)
}
