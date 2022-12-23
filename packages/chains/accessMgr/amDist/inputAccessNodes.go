// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package amDist

import (
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type inputAccessNodes struct {
	chainID     isc.ChainID
	accessNodes []*cryptolib.PublicKey
}

var _ gpa.Input = &inputAccessNodes{}

func NewInputAccessNodes(chainID isc.ChainID, accessNodes []*cryptolib.PublicKey) gpa.Input {
	return &inputAccessNodes{chainID: chainID, accessNodes: accessNodes}
}
