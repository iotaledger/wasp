// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainmanager

import (
	"fmt"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/cmt_log"
	"github.com/iotaledger/wasp/packages/gpa"
)

type inputConsensusTimeout struct {
	committeeAddr iotago.Ed25519Address
	logIndex      cmt_log.LogIndex
}

func NewInputConsensusTimeout(committeeAddr iotago.Ed25519Address, logIndex cmt_log.LogIndex) gpa.Input {
	return &inputConsensusTimeout{
		committeeAddr: committeeAddr,
		logIndex:      logIndex,
	}
}

func (inp *inputConsensusTimeout) String() string {
	return fmt.Sprintf(
		"{chainMgr.inputConsensusTimeout, committeeAddr=%v, logIndex=%v}",
		inp.committeeAddr.String(),
		inp.logIndex,
	)
}
