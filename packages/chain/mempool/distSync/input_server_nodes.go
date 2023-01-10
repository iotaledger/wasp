// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package distSync

import (
	"strings"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/gpa"
)

type inputServerNodes struct {
	serverNodes    []gpa.NodeID
	committeeNodes []gpa.NodeID
}

func NewInputServerNodes(serverNodes, committeeNodes []gpa.NodeID) gpa.Input {
	return &inputServerNodes{serverNodes: serverNodes, committeeNodes: committeeNodes}
}

func (inp *inputServerNodes) String() string {
	str := "serverNodes: "
	str += strings.Join(lo.Map(inp.serverNodes, func(nodeID gpa.NodeID, idx int) string { return nodeID.ShortString() }), ",")
	str += " committeeNodes: "
	str += strings.Join(lo.Map(inp.committeeNodes, func(nodeID gpa.NodeID, idx int) string { return nodeID.ShortString() }), ",")
	return str
}
