// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainMgr

import (
	"github.com/iotaledger/wasp/packages/chain/consensus/journal"
	"github.com/iotaledger/wasp/packages/gpa"
)

type inputConsensusTimeout struct {
	committeeID CommitteeID
	logIndex    journal.LogIndex
}

func NewInputConsensusTimeout(committeeID CommitteeID, logIndex journal.LogIndex) gpa.Input { // TODO: Call it.
	return &inputConsensusTimeout{
		committeeID: committeeID,
		logIndex:    logIndex,
	}
}
