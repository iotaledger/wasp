// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package distsync

import (
	"github.com/iotaledger/wasp/v2/packages/gpa"
)

const (
	msgTypeShareRequest gpa.MessageType = iota
	msgTypeMissingRequest
)

func (dsi *distSyncImpl) UnmarshalPayload(data []byte) (msg gpa.MessagePayload, err error) {
	return gpa.UnmarshalPayload(data, gpa.PayloadAllocator{
		msgTypeMissingRequest: func() gpa.MessagePayload { return new(msgMissingRequest) },
		msgTypeShareRequest:   func() gpa.MessagePayload { return new(msgShareRequest) },
	})
}
