// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package committeelog

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
)

type inputConsensusOutputRejected struct {
	anchor   *isc.StateAnchor
	logIndex LogIndex
}

func NewInputConsensusOutputRejected(anchor *isc.StateAnchor, logIndex LogIndex) gpa.Input {
	return &inputConsensusOutputRejected{
		anchor:   anchor,
		logIndex: logIndex,
	}
}

func (inp *inputConsensusOutputRejected) String() string {
	return fmt.Sprintf("{committeeLog.inputConsensusOutputRejected, %v, li=%v}", inp.anchor, inp.logIndex)
}
