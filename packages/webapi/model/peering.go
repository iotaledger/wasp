// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package model

// PeeringTrustedNode describes single node in the list of trusted peering nodes.
type PeeringTrustedNode struct {
	PubKey string `json:"pubKey" swagger:"desc(Public key of the NetID, base58.)"`
	NetID  string `json:"netID" swagger:"desc(NetID of a peer to trust.)"`
}
