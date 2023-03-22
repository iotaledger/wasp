// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtLog

import (
	"fmt"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type inputConsensusOutputDone struct {
	logIndex          LogIndex
	proposedBaseAO    iotago.OutputID        // Proposed BaseAO
	baseAliasOutputID iotago.OutputID        // Decided BaseAO
	nextAliasOutput   *isc.AliasOutputWithID // And the next one.
}

// This message is internal one, but should be sent by other components (e.g. consensus or the chain).
func NewInputConsensusOutputDone(
	logIndex LogIndex,
	proposedBaseAO iotago.OutputID,
	baseAliasOutputID iotago.OutputID,
	nextAliasOutput *isc.AliasOutputWithID,
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
