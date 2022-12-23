// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package amDist

import (
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
)

type inputTrustedNodes struct {
	trustedNodes []*cryptolib.PublicKey
}

var _ gpa.Input = &inputTrustedNodes{}

func NewInputTrustedNodes(trustedNodes []*cryptolib.PublicKey) gpa.Input {
	return &inputTrustedNodes{trustedNodes: trustedNodes}
}
