// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dist

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/gpa"
)

const (
	msgTypeAccess gpa.MessageType = iota
)

func (amd *AccessMgrDist) MarshalPayload(payload any) ([]byte, error) {
	switch p := payload.(type) {
	case msgAccess:
		return gpa.MarshalPayload(msgTypeAccess, p)
	default:
		panic(fmt.Errorf("accessMgrDist: unknown payload type %T", payload))
	}
}

func (amd *AccessMgrDist) UnmarshalPayload(data []byte) (any, error) {
	return gpa.UnmarshalPayload(data, gpa.PayloadAllocator{
		msgTypeAccess: func() any { return msgAccess{} },
	})
}
