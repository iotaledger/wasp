// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package distsync

import "github.com/iotaledger/wasp/v2/packages/gpa"

type inputTimeTick struct{}

func NewInputTimeTick() gpa.Input {
	return &inputTimeTick{}
}
