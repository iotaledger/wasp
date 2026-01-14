// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package distsync

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/gpa"
)

const (
	msgTypeShareRequest gpa.MessageType = iota
	msgTypeMissingRequest
)

func (dsi *DistSync) MarshalPayload(payload any) (data []byte, err error) {
	switch p := payload.(type) {
	case msgMissingRequest:
		return gpa.MarshalPayload(msgTypeMissingRequest, p)
	case msgShareRequest:
		return gpa.MarshalPayload(msgTypeShareRequest, p)
	default:
		panic(fmt.Errorf("distSync: unknown payload type %T", payload))
	}
}

func (dsi *DistSync) UnmarshalPayload(data []byte) (msg any, err error) {
	return gpa.UnmarshalPayload(data, gpa.PayloadAllocator{
		msgTypeMissingRequest: func() any { return new(msgMissingRequest) },
		msgTypeShareRequest:   func() any { return new(msgShareRequest) },
	})
}
