package dto

type ChainNodeStatus struct {
	Node         PeeringNodeStatus
	ForCommittee bool
	ForAccess    bool
	AccessAPI    string
}

type ChainNodeInfo struct {
	CommitteeNodes []*ChainNodeStatus
	AccessNodes    []*ChainNodeStatus
	CandidateNodes []*ChainNodeStatus
}
