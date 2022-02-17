// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"github.com/iotaledger/wasp/packages/peering"
)

// PeeringTrustedNode describes single node in the list of trusted peering nodes.
type PeeringTrustedNode struct {
	PubKey string `json:"pubKey" swagger:"desc(Public key of the NetID.)"`
	NetID  string `json:"netID" swagger:"desc(NetID of a peer to trust.)"`
}

func NewPeeringTrustedNode(tp *peering.TrustedPeer) *PeeringTrustedNode {
	return &PeeringTrustedNode{
		PubKey: tp.PubKey.AsString(),
		NetID:  tp.NetID,
	}
}

type PeeringNodeStatus struct {
	PubKey   string
	NetID    string
	IsAlive  bool
	NumUsers int
}
