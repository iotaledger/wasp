package dto

import "github.com/iotaledger/wasp/v2/packages/cryptolib"

type ChainNodeStatus struct {
	AccessAPI    string
	ForAccess    bool
	ForCommittee bool
	Node         PeeringNodeStatus
}

type ChainNodeInfo struct {
	Address        *cryptolib.Address
	AccessNodes    []*ChainNodeStatus
	CandidateNodes []*ChainNodeStatus
	CommitteeNodes []*ChainNodeStatus
}
