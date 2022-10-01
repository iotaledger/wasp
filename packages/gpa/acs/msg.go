// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package acs

import "github.com/iotaledger/wasp/packages/gpa"

const (
	msgTypeWrapped byte = iota
)

func (a *acsImpl) UnmarshalMessage(data []byte) (gpa.Message, error) {
	return a.msgWrapper.UnmarshalMessage(data)
}
