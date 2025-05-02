// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtlog

import (
	"github.com/iotaledger/wasp/packages/gpa"
)

const (
	msgTypeNextLogIndex gpa.MessageType = iota
)

func (cl *cmtLogImpl) UnmarshalMessage(data []byte) (gpa.Message, error) {
	return UnmarshalMessage(data)
}

func UnmarshalMessage(data []byte) (gpa.Message, error) {
	return gpa.UnmarshalMessage(data, gpa.Mapper{
		msgTypeNextLogIndex: func() gpa.Message { return new(MsgNextLogIndex) },
	})
}
