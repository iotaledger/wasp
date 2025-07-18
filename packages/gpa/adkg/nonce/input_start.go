// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package nonce

import (
	"github.com/iotaledger/wasp/v2/packages/gpa"
)

type inputStart struct{}

func NewInputStart() gpa.Input {
	return &inputStart{}
}
