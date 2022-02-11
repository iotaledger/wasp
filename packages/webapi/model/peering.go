// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/mr-tron/base58"
)

// PeeringTrustedNode describes single node in the list of trusted peering nodes.
type PeeringTrustedNode struct {
	PubKey string `json:"pubKey" swagger:"desc(Public key of the NetID, base58.)"`
	NetID  string `json:"netID" swagger:"desc(NetID of a peer to trust.)"`
}

func NewPeeringTrustedNode(tp *peering.TrustedPeer) *PeeringTrustedNode {
	return &PeeringTrustedNode{
		PubKey: base58.Encode(tp.PubKey[:]),
		NetID:  tp.NetID,
	}
}

type PeeringNodeStatus struct {
	PubKey   string
	NetID    string
	IsAlive  bool
	NumUsers int
}
