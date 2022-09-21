package dto

type ChainNodeStatus struct {
	AccessAPI    string
	ForAccess    bool
	ForCommittee bool
	Node         PeeringNodeStatus
}

type ChainNodeInfo struct {
	AccessNodes    []*ChainNodeStatus
	CandidateNodes []*ChainNodeStatus
	CommitteeNodes []*ChainNodeStatus
}
