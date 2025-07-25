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

func (dsi *distSyncImpl) UnmarshalMessage(data []byte) (msg gpa.Message, err error) {
	return gpa.UnmarshalMessage(data, gpa.Mapper{
		msgTypeMissingRequest: func() gpa.Message { return new(msgMissingRequest) },
		msgTypeShareRequest:   func() gpa.Message { return new(msgShareRequest) },
	})
}
