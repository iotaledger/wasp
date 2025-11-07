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

func (cmi *ChainMgr) UnmarshalMessage(data []byte) (gpa.Message, error) {
	return gpa.UnmarshalMessage(data, gpa.Mapper{
		msgTypeMsgNextLogIndex: func() gpa.Message { return new(msgNextLogIndex) },
		msgTypeBlockProduced:   func() gpa.Message { return new(msgBlockProduced) },
	})
}
