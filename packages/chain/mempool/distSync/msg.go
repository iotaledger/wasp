// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package distSync

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
)

const (
	msgTypeShareRequest byte = iota
	msgTypeMissingRequest
)

func (dsi *distSyncImpl) UnmarshalMessage(data []byte) (msg gpa.Message, err error) {
	switch data[0] {
	case msgTypeMissingRequest:
		msg = &msgMissingRequest{}
	case msgTypeShareRequest:
		msg = &msgShareRequest{}
	default:
		return nil, fmt.Errorf("unknown message type %b", data[0])
	}
	err = msg.UnmarshalBinary(data)
	if err != nil {
		return nil, err
	}
	return msg, nil
}
