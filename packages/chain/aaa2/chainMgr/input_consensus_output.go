// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainMgr

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/consensus/journal"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type inputConsensusOutput struct {
	committeeID       CommitteeID
	logIndex          journal.LogIndex
	baseAliasOutputID iotago.OutputID
	nextAliasOutput   *isc.AliasOutputWithID
}

func NewInputConsensusOutput(
	committeeID CommitteeID,
	logIndex journal.LogIndex,
	baseAliasOutputID iotago.OutputID,
	nextAliasOutput *isc.AliasOutputWithID,
) gpa.Input {
	return &inputConsensusOutput{
		committeeID:       committeeID,
		logIndex:          logIndex,
		baseAliasOutputID: baseAliasOutputID,
		nextAliasOutput:   nextAliasOutput,
	}
}
