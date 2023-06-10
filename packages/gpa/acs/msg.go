// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package acs

import (
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

const (
	msgTypeWrapped rwutil.Kind = iota
)

func (a *acsImpl) UnmarshalMessage(data []byte) (gpa.Message, error) {
	return a.msgWrapper.UnmarshalMessage(data)
}
