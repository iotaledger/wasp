// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package committeelog

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
)

type inputConsensusOutputConfirmed struct {
	nextAnchor *isc.StateAnchor
	logIndex   LogIndex
}

func NewInputConsensusOutputConfirmed(nextAnchor *isc.StateAnchor, logIndex LogIndex) gpa.Input {
	return &inputConsensusOutputConfirmed{
		nextAnchor: nextAnchor,
		logIndex:   logIndex,
	}
}

func (inp *inputConsensusOutputConfirmed) String() string {
	return fmt.Sprintf("{committeeLog.inputConsensusOutputConfirmed, result=%v, li=%v}", inp.nextAnchor, inp.logIndex)
}
