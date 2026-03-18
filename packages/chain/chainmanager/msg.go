// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainmanager

import (
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/state"
)

const (
	msgTypeCommitteeLog gpa.MessageType = iota
	msgTypeBlockProduced
)

func (cmi *ChainMgr) UnmarshalPayload(data []byte) (gpa.MessagePayload, error) {
	return gpa.UnmarshalPayload(data, gpa.PayloadAllocator{
		msgTypeCommitteeLog: func() gpa.MessagePayload { return new(msgCommitteeLog) },
		msgTypeBlockProduced: func() gpa.MessagePayload {
			msgBlock := new(msgBlockProduced)

			// TODO: Validate if we ever have different block implementations.
			msgBlock.block = state.NewBlock()

			return msgBlock
		},
	})
}
