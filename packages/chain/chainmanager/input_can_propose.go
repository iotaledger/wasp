// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainmanager

import (
	"github.com/iotaledger/wasp/v2/packages/gpa"
)

type inputCanPropose struct{}

func NewInputCanPropose() gpa.Input {
	return &inputCanPropose{}
}

func (inp *inputCanPropose) String() string {
	return "{chainMgr.inputCanPropose}"
}
