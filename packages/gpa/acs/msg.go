// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package acs

import (
	"github.com/iotaledger/wasp/v2/packages/gpa"
)

const (
	msgTypeWrapped gpa.MessageType = iota
)

func (a *ACS) UnmarshalPayload(data []byte) (gpa.MessagePayload, error) {
	return gpa.UnmarshalPayload(data, gpa.PayloadAllocator{}, gpa.PayloadFallback{
		msgTypeWrapped: a.msgWrapper.UnmarshalPayload,
	})
}
