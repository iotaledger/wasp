// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainMgr

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/aaa2/cmtLog"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
)

type inputConsensusOutputDone struct {
	committeeAddr     iotago.Ed25519Address
	logIndex          cmtLog.LogIndex
	baseAliasOutputID iotago.OutputID
	nextAliasOutput   *isc.AliasOutputWithID
	nextVirtualState  state.VirtualStateAccess
	transaction       *iotago.Transaction
}

func NewInputConsensusOutputDone(
	committeeAddr iotago.Ed25519Address,
	logIndex cmtLog.LogIndex,
	baseAliasOutputID iotago.OutputID,
	nextAliasOutput *isc.AliasOutputWithID,
	nextVirtualState state.VirtualStateAccess,
	transaction *iotago.Transaction,
) gpa.Input {
	return &inputConsensusOutputDone{
		committeeAddr:     committeeAddr,
		logIndex:          logIndex,
		baseAliasOutputID: baseAliasOutputID,
		nextAliasOutput:   nextAliasOutput,
		nextVirtualState:  nextVirtualState,
		transaction:       transaction,
	}
}
