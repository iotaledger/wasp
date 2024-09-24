// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package distsync

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
)

const (
	msgTypeShareRequest gpa.MessageType = iota
	msgTypeMissingRequest
)

func (dsi *distSyncImpl) MarshalMessage(msg gpa.Message) ([]byte, error) {
	switch msg := msg.(type) {
	case *msgShareRequest:
		return gpa.MarshalMessage(msgTypeShareRequest, msg)
	case *msgMissingRequest:
		return gpa.MarshalMessage(msgTypeMissingRequest, msg)
	default:
		return nil, fmt.Errorf("unknown message type for %T: %T", dsi, msg)
	}
}

func (dsi *distSyncImpl) UnmarshalMessage(data []byte) (msg gpa.Message, err error) {
	return gpa.UnmarshalMessage(data, gpa.Mapper{
		msgTypeMissingRequest: func() gpa.Message { return new(msgMissingRequest) },
		msgTypeShareRequest:   func() gpa.Message { return new(msgShareRequest) },
	})
}
