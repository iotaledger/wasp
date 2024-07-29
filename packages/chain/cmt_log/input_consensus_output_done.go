// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmt_log

import (
	"fmt"

	"github.com/iotaledger/wasp/clients/iscmove/iscmove_types"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/sui-go/sui"
)

type inputConsensusOutputDone struct {
	logIndex          LogIndex
	proposedBaseAO    sui.ObjectID          // Proposed BaseAO
	baseAliasOutputID sui.ObjectID          // Decided BaseAO
	nextAliasOutput   *iscmove_types.Anchor // And the next one.
}

// This message is internal one, but should be sent by other components (e.g. consensus or the chain).
func NewInputConsensusOutputDone(
	logIndex LogIndex,
	proposedBaseAO sui.ObjectID,
	baseAliasOutputID sui.ObjectID,
	nextAliasOutput *iscmove_types.Anchor,
) gpa.Input {
	return &inputConsensusOutputDone{
		logIndex:          logIndex,
		proposedBaseAO:    proposedBaseAO,
		baseAliasOutputID: baseAliasOutputID,
		nextAliasOutput:   nextAliasOutput,
	}
}

func (inp *inputConsensusOutputDone) String() string {
	return fmt.Sprintf(
		"{cmtLog.inputConsensusOutputDone, logIndex=%v, proposedBaseAO=%s, baseAliasOutputID=%s, nextAliasOutput=%v}",
		inp.logIndex, inp.proposedBaseAO.String(), inp.baseAliasOutputID.String(), inp.nextAliasOutput,
	)
}
