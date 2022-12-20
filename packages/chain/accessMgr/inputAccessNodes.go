// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package accessMgr

import (
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
)

type inputAccessNodes struct {
	accessNodes []*cryptolib.PublicKey
}

var _ gpa.Input = &inputAccessNodes{}

func NewInputAccessNodes(accessNodes []*cryptolib.PublicKey) gpa.Input {
	return &inputAccessNodes{accessNodes: accessNodes}
}
