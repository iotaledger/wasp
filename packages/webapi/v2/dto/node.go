package dto

import iotago "github.com/iotaledger/iota.go/v3"

type ChainNodeStatus struct {
	AccessAPI    string
	ForAccess    bool
	ForCommittee bool
	Node         PeeringNodeStatus
}

type ChainNodeInfo struct {
	Address        iotago.Address
	AccessNodes    []*ChainNodeStatus
	CandidateNodes []*ChainNodeStatus
	CommitteeNodes []*ChainNodeStatus
}
