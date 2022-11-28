// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtLog

import (
	"github.com/iotaledger/wasp/packages/gpa"
)

// Sent by the Chain, when another committee has received control over this chain.
type inputSuspend struct{}

func NewInputSuspend() gpa.Input {
	return &inputSuspend{}
}
