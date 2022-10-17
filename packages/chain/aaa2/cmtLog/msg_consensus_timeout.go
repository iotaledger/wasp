// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtLog

import (
	"github.com/iotaledger/wasp/packages/chain/consensus/journal"
	"github.com/iotaledger/wasp/packages/gpa"
)

type msgConsensusTimeout struct {
	gpa.BasicMessage
	logIndex journal.LogIndex
}

var _ gpa.Message = &msgConsensusTimeout{}

// This message is internal one, but should be sent by other components (e.g. consensus or the chain).
func NewMsgConsensusTimeout(
	recipient gpa.NodeID,
	logIndex journal.LogIndex,
) gpa.Message {
	return &msgConsensusTimeout{
		BasicMessage: gpa.NewBasicMessage(recipient),
		logIndex:     logIndex,
	}
}

func (m *msgConsensusTimeout) MarshalBinary() ([]byte, error) {
	panic("trying to marshal local message")
}
