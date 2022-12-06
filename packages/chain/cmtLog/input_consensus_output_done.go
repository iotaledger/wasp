// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtLog

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type inputConsensusOutputDone struct {
	logIndex          LogIndex
	baseAliasOutputID iotago.OutputID
	nextAliasOutput   *isc.AliasOutputWithID
}

// This message is internal one, but should be sent by other components (e.g. consensus or the chain).
func NewInputConsensusOutputDone(
	logIndex LogIndex,
	baseAliasOutputID iotago.OutputID,
	nextAliasOutput *isc.AliasOutputWithID,
) gpa.Input {
	return &inputConsensusOutputDone{
		logIndex:          logIndex,
		baseAliasOutputID: baseAliasOutputID,
		nextAliasOutput:   nextAliasOutput,
	}
}
