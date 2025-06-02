// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainmanager

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/chain/cmtlog"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
)

type inputConsensusOutputSkip struct {
	committeeAddr cryptolib.Address
	logIndex      cmtlog.LogIndex
}

func NewInputConsensusOutputSkip(
	committeeAddr cryptolib.Address,
	logIndex cmtlog.LogIndex,
) gpa.Input {
	return &inputConsensusOutputSkip{
		committeeAddr: committeeAddr,
		logIndex:      logIndex,
	}
}

func (inp *inputConsensusOutputSkip) String() string {
	return fmt.Sprintf(
		"{chainMgr.inputConsensusOutputSkip, committeeAddr=%v, logIndex=%v}",
		inp.committeeAddr.String(),
		inp.logIndex,
	)
}
