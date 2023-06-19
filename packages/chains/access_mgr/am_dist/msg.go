// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package am_dist

import (
	"github.com/iotaledger/wasp/packages/gpa"
)

const (
	msgTypeAccess gpa.MessageType = iota
)

func (amd *accessMgrDist) UnmarshalMessage(data []byte) (gpa.Message, error) {
	return gpa.UnmarshalMessage(data, gpa.Mapper{
		msgTypeAccess: func() gpa.Message { return new(msgAccess) },
	})
}
