// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainmanager

import (
	"fmt"

	iotago "github.com/iotaledger/iota.go/v4"
	"github.com/iotaledger/wasp/packages/chain/cmt_log"
	"github.com/iotaledger/wasp/packages/chain/cons"
	"github.com/iotaledger/wasp/packages/gpa"
)

type inputConsensusOutputDone struct {
	committeeAddr   iotago.Ed25519Address
	logIndex        cmt_log.LogIndex
	consensusResult *cons.Result
}

func NewInputConsensusOutputDone(
	committeeAddr iotago.Ed25519Address,
	logIndex cmt_log.LogIndex,
	consensusResult *cons.Result,
) gpa.Input {
	return &inputConsensusOutputDone{
		committeeAddr:   committeeAddr,
		logIndex:        logIndex,
		consensusResult: consensusResult,
	}
}

func (inp *inputConsensusOutputDone) String() string {
	return fmt.Sprintf(
		"{chainMgr.inputConsensusOutputDone, committeeAddr=%v, logIndex=%v, consensusResult=%v}",
		inp.committeeAddr.String(),
		inp.logIndex,
		inp.consensusResult,
	)
}
