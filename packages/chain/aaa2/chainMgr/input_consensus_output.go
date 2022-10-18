// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainMgr

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/aaa2/cmtLog"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type inputConsensusOutput struct {
	committeeID       CommitteeID
	logIndex          cmtLog.LogIndex
	baseAliasOutputID iotago.OutputID
	nextAliasOutput   *isc.AliasOutputWithID
	transaction       *iotago.Transaction
}

func NewInputConsensusOutput(
	committeeID CommitteeID,
	logIndex cmtLog.LogIndex,
	baseAliasOutputID iotago.OutputID,
	nextAliasOutput *isc.AliasOutputWithID,
	transaction *iotago.Transaction,
) gpa.Input {
	return &inputConsensusOutput{
		committeeID:       committeeID,
		logIndex:          logIndex,
		baseAliasOutputID: baseAliasOutputID,
		nextAliasOutput:   nextAliasOutput,
		transaction:       transaction,
	}
}
