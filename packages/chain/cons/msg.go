// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"github.com/iotaledger/wasp/v2/packages/gpa"
)

const (
	msgTypeBLSShare gpa.MessageType = iota
	msgTypeWrapped
)

func (c *consImpl) UnmarshalMessage(data []byte) (gpa.Message, error) {
	return gpa.UnmarshalMessage(data, gpa.Mapper{
		msgTypeBLSShare: func() gpa.Message { return &msgBLSPartialSig{blsSuite: c.blsSuite} },
	}, gpa.Fallback{
		msgTypeWrapped: c.msgWrapper.UnmarshalMessage,
	})
}
