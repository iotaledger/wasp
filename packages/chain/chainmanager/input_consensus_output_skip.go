// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainmanager

import (
	"fmt"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/chain/cmt_log"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
)

type inputConsensusOutputSkip struct {
	committeeAddr  cryptolib.Address
	logIndex       cmt_log.LogIndex
	proposedBaseAO *iotago.ObjectRef
}

func NewInputConsensusOutputSkip(
	committeeAddr cryptolib.Address,
	logIndex cmt_log.LogIndex,
	proposedBaseAO *iotago.ObjectRef,
) gpa.Input {
	return &inputConsensusOutputSkip{
		committeeAddr:  committeeAddr,
		logIndex:       logIndex,
		proposedBaseAO: proposedBaseAO,
	}
}

func (inp *inputConsensusOutputSkip) String() string {
	return fmt.Sprintf(
		"{chainMgr.inputConsensusOutputSkip, committeeAddr=%v, logIndex=%v, proposedBaseAO=%s}",
		inp.committeeAddr.String(),
		inp.logIndex,
		inp.proposedBaseAO.String(),
	)
}
