// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dss

import (
	"github.com/iotaledger/wasp/packages/gpa"
)

type inputDecided struct {
	decidedIndexProposals map[gpa.NodeID][]int
	messageToSign         []byte
}

func NewInputDecided(decidedIndexProposals map[gpa.NodeID][]int, messageToSign []byte) gpa.Input {
	return &inputDecided{
		decidedIndexProposals: decidedIndexProposals,
		messageToSign:         messageToSign,
	}
}
