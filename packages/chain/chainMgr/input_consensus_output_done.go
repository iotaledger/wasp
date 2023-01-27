// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainMgr

import (
	"fmt"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/cmtLog"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
)

type inputConsensusOutputDone struct {
	committeeAddr     iotago.Ed25519Address
	logIndex          cmtLog.LogIndex
	baseAliasOutputID iotago.OutputID
	nextAliasOutput   *isc.AliasOutputWithID
	nextState         state.StateDraft
	transaction       *iotago.Transaction
}

func NewInputConsensusOutputDone(
	committeeAddr iotago.Ed25519Address,
	logIndex cmtLog.LogIndex,
	baseAliasOutputID iotago.OutputID,
	nextAliasOutput *isc.AliasOutputWithID,
	nextState state.StateDraft,
	transaction *iotago.Transaction,
) gpa.Input {
	return &inputConsensusOutputDone{
		committeeAddr:     committeeAddr,
		logIndex:          logIndex,
		baseAliasOutputID: baseAliasOutputID,
		nextAliasOutput:   nextAliasOutput,
		nextState:         nextState,
		transaction:       transaction,
	}
}

func (inp *inputConsensusOutputDone) String() string {
	txID, err := inp.transaction.ID()
	if err != nil {
		txID = iotago.TransactionID{}
	}
	return fmt.Sprintf(
		"{chainMgr.inputConsensusOutputDone, committeeAddr=%v, logIndex=%v, baseAliasOutputID=%v, nextAliasOutput.ID=%v, transaction=%v}",
		inp.committeeAddr.String(),
		inp.logIndex,
		inp.baseAliasOutputID.ToHex(),
		inp.nextAliasOutput.OutputID().ToHex(),
		txID.ToHex(),
	)
}
