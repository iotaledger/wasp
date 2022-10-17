// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainMgr

import (
	"github.com/iotaledger/wasp/packages/chain/consensus/journal"
	"github.com/iotaledger/wasp/packages/gpa"
)

type msgConsensusTimeout struct {
	gpa.BasicMessage
	committeeID CommitteeID
	logIndex    journal.LogIndex
}

var _ gpa.Message = &msgConsensusTimeout{}

func NewMsgConsensusTimeout(recipient gpa.NodeID, committeeID CommitteeID, logIndex journal.LogIndex) gpa.Message { // TODO: Call it.
	return &msgConsensusTimeout{
		BasicMessage: gpa.NewBasicMessage(recipient),
		committeeID:  committeeID,
		logIndex:     logIndex,
	}
}

func (m *msgConsensusTimeout) MarshalBinary() ([]byte, error) {
	panic("that's local message")
}
