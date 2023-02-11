// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dss

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/hashing"
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

func (inp *inputDecided) String() string {
	var msgHash *hashing.HashValue
	if inp.messageToSign != nil {
		msgHashVal := hashing.HashDataBlake2b(inp.messageToSign)
		msgHash = &msgHashVal
	}
	return fmt.Sprintf("{chain.dss.inputDecided, decidedIndexProposals=%+v, H(messageToSign)=%v}", inp.decidedIndexProposals, msgHash)
}
