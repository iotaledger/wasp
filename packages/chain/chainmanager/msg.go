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

func (chainMgr *chainMgrImpl) UnmarshalMessage(data []byte) (gpa.Message, error) {
	return gpa.UnmarshalMessage(data, gpa.Mapper{
		msgTypeCommitteeLog: func() gpa.Message { return new(msgCommitteeLog) },
		msgTypeBlockProduced: func() gpa.Message {
			msgBlock := new(msgBlockProduced)

			// TODO: Validate if we ever have different block implementations.
			msgBlock.block = state.NewBlock()

			return msgBlock
		},
	})
}
