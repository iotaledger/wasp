// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmt_log

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc/sui"
	"github.com/iotaledger/wasp/sui-go/sui_types"
)

type inputConsensusOutputDone struct {
	logIndex          LogIndex
	proposedBaseAO    sui_types.ObjectID // Proposed BaseAO
	baseAliasOutputID sui_types.ObjectID // Decided BaseAO
	nextAliasOutput   *sui.Anchor        // And the next one.
}

// This message is internal one, but should be sent by other components (e.g. consensus or the chain).
func NewInputConsensusOutputDone(
	logIndex LogIndex,
	proposedBaseAO sui_types.ObjectID,
	baseAliasOutputID sui_types.ObjectID,
	nextAliasOutput *sui.Anchor,
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
		"{cmtLog.inputConsensusOutputDone, logIndex=%v, proposedBaseAO=%v, baseAliasOutputID=%v, nextAliasOutput=%v}",
		inp.logIndex, inp.proposedBaseAO.ToHex(), inp.baseAliasOutputID.ToHex(), inp.nextAliasOutput,
	)
}
