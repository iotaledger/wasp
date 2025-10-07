// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainmanager

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/chain/cmtlog"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
)

type inputConsensusOutputSkip struct {
	committeeAddr cryptolib.Address
	logIndex      cmtlog.LogIndex
}

func NewInputConsensusOutputSkip(
	committeeAddr cryptolib.Address,
	logIndex cmtlog.LogIndex,
) *inputConsensusOutputSkip {
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
