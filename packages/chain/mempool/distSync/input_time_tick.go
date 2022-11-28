// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package distSync

import "github.com/iotaledger/wasp/packages/gpa"

type inputTimeTick struct{}

func NewInputTimeTick() gpa.Input {
	return &inputTimeTick{}
}
