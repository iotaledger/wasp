// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"github.com/iotaledger/wasp/v2/packages/gpa"
)

const (
	msgTypeBLSShare gpa.MessageType = iota
	msgTypeWrapped
)

func (c *Consensus) UnmarshalPayload(data []byte) (gpa.MessagePayload, error) {
	return gpa.UnmarshalPayload(data, gpa.PayloadAllocator{
		msgTypeBLSShare: func() gpa.MessagePayload { return &msgBLSPartialSig{blsSuite: c.blsSuite} },
	}, gpa.PayloadFallback{
		msgTypeWrapped: c.msgWrapper.UnmarshalPayload,
	})
}
