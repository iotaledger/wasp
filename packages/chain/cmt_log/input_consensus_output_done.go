// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmt_log

import (
	"fmt"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/gpa"
)

type inputConsensusOutputDone struct {
	logIndex       LogIndex
	proposedBaseAO iotago.ObjectID   // Proposed BaseAO
	baseAnchorRef  *iotago.ObjectRef // Decided BaseAO
}

// This message is internal one, but should be sent by other components (e.g. consensus or the chain).
func NewInputConsensusOutputDone(
	logIndex LogIndex,
	proposedBaseAO iotago.ObjectID,
	baseAnchorRef *iotago.ObjectRef,
) gpa.Input {
	return &inputConsensusOutputDone{
		logIndex:       logIndex,
		proposedBaseAO: proposedBaseAO,
		baseAnchorRef:  baseAnchorRef,
	}
}

func (inp *inputConsensusOutputDone) String() string {
	return fmt.Sprintf(
		"{cmtLog.inputConsensusOutputDone, logIndex=%v, proposedBaseAO=%s, baseAnchorRef=%s}",
		inp.logIndex, inp.proposedBaseAO.String(), inp.baseAnchorRef.String(),
	)
}
