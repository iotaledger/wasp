// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package acs

import (
	"github.com/iotaledger/wasp/v2/packages/gpa"
)

const (
	msgTypeWrapped gpa.MessageType = iota
)

func (a *acsImpl) UnmarshalMessage(data []byte) (gpa.Message, error) {
	return gpa.UnmarshalMessage(data, gpa.Mapper{}, gpa.Fallback{
		msgTypeWrapped: a.msgWrapper.UnmarshalMessage,
	})
}
