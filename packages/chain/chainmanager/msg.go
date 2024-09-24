// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainmanager

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
)

const (
	msgTypeCmtLog gpa.MessageType = iota
	msgTypeBlockProduced
)

func (cmi *chainMgrImpl) MarshalMessage(msg gpa.Message) ([]byte, error) {
	switch msg := msg.(type) {
	case *msgCmtLog:
		return gpa.MarshalMessage(msgTypeCmtLog, msg)
	case *msgBlockProduced:
		return gpa.MarshalMessage(msgTypeBlockProduced, msg)
	default:
		return nil, fmt.Errorf("unknown message type for %T: %T", cmi, msg)
	}
}

func (cmi *chainMgrImpl) UnmarshalMessage(data []byte) (gpa.Message, error) {
	return gpa.UnmarshalMessage(data, gpa.Mapper{
		msgTypeCmtLog:        func() gpa.Message { return new(msgCmtLog) },
		msgTypeBlockProduced: func() gpa.Message { return new(msgBlockProduced) },
	})
}
