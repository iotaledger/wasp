// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainMgr

import (
	"github.com/iotaledger/wasp/packages/chain/aaa2/cmtLog"
	"github.com/iotaledger/wasp/packages/gpa"
)

type inputConsensusTimeout struct {
	committeeID CommitteeID
	logIndex    cmtLog.LogIndex
}

func NewInputConsensusTimeout(committeeID CommitteeID, logIndex cmtLog.LogIndex) gpa.Input { // TODO: Call it.
	return &inputConsensusTimeout{
		committeeID: committeeID,
		logIndex:    logIndex,
	}
}
