// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtLog

import (
	"github.com/iotaledger/wasp/packages/gpa"
)

type msgConsensusTimeout struct {
	gpa.BasicMessage
	logIndex LogIndex
}

var _ gpa.Message = &msgConsensusTimeout{}

// This message is internal one, but should be sent by other components (e.g. consensus or the chain).
func NewMsgConsensusTimeout(
	recipient gpa.NodeID,
	logIndex LogIndex,
) gpa.Message {
	return &msgConsensusTimeout{
		BasicMessage: gpa.NewBasicMessage(recipient),
		logIndex:     logIndex,
	}
}

func (m *msgConsensusTimeout) MarshalBinary() ([]byte, error) {
	panic("trying to marshal local message")
}
