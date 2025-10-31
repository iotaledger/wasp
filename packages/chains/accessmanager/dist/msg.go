// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dist

import (
	"github.com/iotaledger/wasp/v2/packages/gpa"
)

const (
	msgTypeAccess gpa.MessageType = iota
)

func (amd *accessMgrDist) UnmarshalPayload(data []byte) (gpa.MessagePayload, error) {
	return gpa.UnmarshalPayload(data, gpa.PayloadAllocator{
		msgTypeAccess: func() gpa.MessagePayload { return new(msgAccess) },
	})
}
