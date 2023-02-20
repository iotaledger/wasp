// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainMgr

import (
	"fmt"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/cmtLog"
	"github.com/iotaledger/wasp/packages/chain/cons"
	"github.com/iotaledger/wasp/packages/gpa"
)

type inputConsensusOutputDone struct {
	committeeAddr   iotago.Ed25519Address
	logIndex        cmtLog.LogIndex
	proposedBaseAO  iotago.OutputID
	consensusResult *cons.Result
}

func NewInputConsensusOutputDone(
	committeeAddr iotago.Ed25519Address,
	logIndex cmtLog.LogIndex,
	proposedBaseAO iotago.OutputID,
	consensusResult *cons.Result,
) gpa.Input {
	return &inputConsensusOutputDone{
		committeeAddr:   committeeAddr,
		logIndex:        logIndex,
		proposedBaseAO:  proposedBaseAO,
		consensusResult: consensusResult,
	}
}

func (inp *inputConsensusOutputDone) String() string {
	return fmt.Sprintf(
		"{chainMgr.inputConsensusOutputDone, committeeAddr=%v, logIndex=%v, proposedBaseAO=%v, consensusResult=%v}",
		inp.committeeAddr.String(),
		inp.logIndex,
		inp.proposedBaseAO.ToHex(),
		inp.consensusResult,
	)
}
