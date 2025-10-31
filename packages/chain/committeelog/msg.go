// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package committeelog

import (
	"github.com/iotaledger/wasp/v2/packages/gpa"
)

const (
	msgTypeNextLogIndex gpa.MessageType = iota
)

func (cl *CommitteeLog) UnmarshalPayload(data []byte) (gpa.MessagePayload, error) {
	return UnmarshalPayload(data)
}

func UnmarshalPayload(data []byte) (gpa.MessagePayload, error) {
	return gpa.UnmarshalPayload(data, gpa.PayloadAllocator{
		msgTypeNextLogIndex: func() gpa.MessagePayload { return new(MsgNextLogIndex) },
	})
}
