// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package distSync

import (
	"github.com/iotaledger/wasp/packages/gpa"
)

type inputAccessNodes struct {
	accessNodes    []gpa.NodeID
	committeeNodes []gpa.NodeID
}

func NewInputAccessNodes(accessNodes, committeeNodes []gpa.NodeID) gpa.Input {
	return &inputAccessNodes{accessNodes: accessNodes, committeeNodes: committeeNodes}
}
