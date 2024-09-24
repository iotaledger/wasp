// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmt_log

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
)

const (
	msgTypeNextLogIndex gpa.MessageType = iota
)

func (cl *cmtLogImpl) MarshalMessage(msg gpa.Message) ([]byte, error) {
	return MarshalMessage(msg)
}

func (cl *cmtLogImpl) UnmarshalMessage(data []byte) (gpa.Message, error) {
	return UnmarshalMessage(data)
}

func MarshalMessage(msg gpa.Message) ([]byte, error) {
	switch msg := msg.(type) {
	case *MsgNextLogIndex:
		return gpa.MarshalMessage(msgTypeNextLogIndex, msg)
	default:
		return nil, fmt.Errorf("unknown cmt_log message type: %T", msg)
	}
}

func UnmarshalMessage(data []byte) (gpa.Message, error) {
	return gpa.UnmarshalMessage(data, gpa.Mapper{
		msgTypeNextLogIndex: func() gpa.Message { return new(MsgNextLogIndex) },
	})
}
