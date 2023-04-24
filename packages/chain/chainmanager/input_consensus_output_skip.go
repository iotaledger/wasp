// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainmanager

import (
	"fmt"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/cmt_log"
	"github.com/iotaledger/wasp/packages/gpa"
)

type inputConsensusOutputSkip struct {
	committeeAddr  iotago.Ed25519Address
	logIndex       cmt_log.LogIndex
	proposedBaseAO iotago.OutputID
}

func NewInputConsensusOutputSkip(
	committeeAddr iotago.Ed25519Address,
	logIndex cmt_log.LogIndex,
	proposedBaseAO iotago.OutputID,
) gpa.Input {
	return &inputConsensusOutputSkip{
		committeeAddr:  committeeAddr,
		logIndex:       logIndex,
		proposedBaseAO: proposedBaseAO,
	}
}

func (inp *inputConsensusOutputSkip) String() string {
	return fmt.Sprintf(
		"{chainMgr.inputConsensusOutputSkip, committeeAddr=%v, logIndex=%v, proposedBaseAO=%v}",
		inp.committeeAddr.String(),
		inp.logIndex,
		inp.proposedBaseAO.ToHex(),
	)
}
