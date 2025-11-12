// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainmanager

import (
	"github.com/iotaledger/wasp/v2/packages/gpa"
)

const (
	msgTypeMsgNextLogIndex gpa.MessageType = iota
	msgTypeBlockProduced
)

func (cmi *ChainMgr) UnmarshalPayload(data []byte) (gpa.MessagePayload, error) {
	return gpa.UnmarshalPayload(data, gpa.PayloadAllocator{
		msgTypeMsgNextLogIndex: func() gpa.MessagePayload { return new(msgNextLogIndex) },
		msgTypeBlockProduced:   func() gpa.MessagePayload { return new(msgBlockProduced) },
	})
}
