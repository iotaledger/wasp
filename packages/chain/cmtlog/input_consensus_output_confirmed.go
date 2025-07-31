// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtlog

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
)

type inputConsensusOutputConfirmed struct {
	nextAnchorObject *isc.StateAnchor
	logIndex         LogIndex
}

func NewInputConsensusOutputConfirmed(nextAnchorObject *isc.StateAnchor, logIndex LogIndex) gpa.Input {
	return &inputConsensusOutputConfirmed{
		nextAnchorObject: nextAnchorObject,
		logIndex:         logIndex,
	}
}

func (inp *inputConsensusOutputConfirmed) String() string {
	return fmt.Sprintf("{cmtLog.inputConsensusOutputConfirmed, result=%v, li=%v}", inp.nextAnchorObject, inp.logIndex)
}
