// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package model

type ChainInfo struct {
	ChainID        ChainIDBech32      `swagger:"desc(ChainID)"`
	Active         bool               `swagger:"desc(Whether or not the chain is active)"`
	StateAddress   Address            `swagger:"desc(State address, if we are part of it.)"`
	CommitteeNodes []*ChainNodeStatus `swagger:"desc(Committee nodes and their peering info.)"`
	AccessNodes    []*ChainNodeStatus `swagger:"desc(Access nodes and their peering info.)"`
	CandidateNodes []*ChainNodeStatus `swagger:"desc(Candidate nodes and their peering info.)"`
}

type ChainNodeStatus struct {
	Node         PeeringNodeStatus
	ForCommittee bool
	ForAccess    bool
	AccessAPI    string
}
