// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package distSync

import (
	"strings"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/gpa"
)

type inputAccessNodes struct {
	accessNodes    []gpa.NodeID
	committeeNodes []gpa.NodeID
}

func NewInputAccessNodes(accessNodes, committeeNodes []gpa.NodeID) gpa.Input {
	return &inputAccessNodes{accessNodes: accessNodes, committeeNodes: committeeNodes}
}

func (inp *inputAccessNodes) String() string {
	str := "accessNodes: "
	str += strings.Join(lo.Map(inp.accessNodes, func(nodeID gpa.NodeID, idx int) string { return nodeID.ShortString() }), ",")
	str += " committeeNodes: "
	str += strings.Join(lo.Map(inp.committeeNodes, func(nodeID gpa.NodeID, idx int) string { return nodeID.ShortString() }), ",")
	return str
}
