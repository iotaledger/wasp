// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package nonce

import (
	"github.com/iotaledger/wasp/packages/gpa"
)

// An event to self.
type inputAgreementResult struct {
	proposals map[gpa.NodeID][]int
}

func NewInputAgreementResult(proposals map[gpa.NodeID][]int) gpa.Input {
	return &inputAgreementResult{proposals: proposals}
}
