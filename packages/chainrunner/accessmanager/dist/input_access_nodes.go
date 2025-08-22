// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dist

import (
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/gpa"
)

type inputAccessNodes struct {
	accessNodes []*cryptolib.PublicKey
}

var _ gpa.Input = &inputAccessNodes{}

func NewInputAccessNodes(accessNodes []*cryptolib.PublicKey) gpa.Input {
	return &inputAccessNodes{accessNodes: accessNodes}
}
