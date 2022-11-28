// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtLog

import (
	"github.com/iotaledger/wasp/packages/gpa"
)

// Sent by the Chain, to start the committee (to send NextLogIndex etc).
type inputStart struct{}

func NewInputStart() gpa.Input {
	return &inputStart{}
}
