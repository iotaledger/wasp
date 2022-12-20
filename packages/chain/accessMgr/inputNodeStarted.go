// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package accessMgr

import (
	"github.com/iotaledger/wasp/packages/gpa"
)

type inputNodeStarted struct{}

var _ gpa.Input = &inputNodeStarted{}

func NewInputNodeStarted() gpa.Input {
	return &inputNodeStarted{}
}
