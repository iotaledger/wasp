// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"github.com/iotaledger/wasp/packages/gpa"
)

type inputStateMgrBlockSaved struct{}

func NewInputStateMgrBlockSaved() gpa.Input {
	return &inputStateMgrBlockSaved{}
}

func (inp *inputStateMgrBlockSaved) String() string {
	return "{cons.inputStateMgrBlockSaved}"
}
