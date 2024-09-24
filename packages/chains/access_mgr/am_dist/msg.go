// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package am_dist

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
)

const (
	msgTypeAccess gpa.MessageType = iota
)

func (amd *accessMgrDist) MarshalMessage(msg gpa.Message) ([]byte, error) {
	switch msg := msg.(type) {
	case *msgAccess:
		return gpa.MarshalMessage(msgTypeAccess, msg)
	default:
		return nil, fmt.Errorf("unknown message type: %T", msg)
	}
}

func (amd *accessMgrDist) UnmarshalMessage(data []byte) (gpa.Message, error) {
	return gpa.UnmarshalMessage(data, gpa.Mapper{
		msgTypeAccess: func() gpa.Message { return new(msgAccess) },
	})
}
