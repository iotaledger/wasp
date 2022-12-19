// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package model

type ChainInfo struct {
	ChainID        ChainIDBech32      `json:"chainId" swagger:"desc(ChainID bech32)"`
	Active         bool               `json:"active" swagger:"desc(Whether or not the chain is active)"`
	StateAddress   Address            `json:"stateAddress" swagger:"desc(State address, if we are part of it.)"`
	CommitteeNodes []*ChainNodeStatus `json:"committeeNodes" swagger:"desc(Committee nodes and their peering info.)"`
	AccessNodes    []*ChainNodeStatus `json:"accessNodes" swagger:"desc(Access nodes and their peering info.)"`
	CandidateNodes []*ChainNodeStatus `json:"candidateNodes" swagger:"desc(Candidate nodes and their peering info.)"`
}

type ChainNodeStatus struct {
	Node         PeeringNodeStatus
	ForCommittee bool
	ForAccess    bool
	AccessAPI    string
}
