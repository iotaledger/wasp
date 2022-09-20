package models

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/webapi/v2/dto"
)

type ChainInfo struct {
	ChainID        *isc.ChainID            `swagger:"desc(ChainID (hex-encoded))"`
	Active         bool               `swagger:"desc(Whether or not the chain is active)"`
	StateAddress   iotago.Address            `swagger:"desc(State address, if we are part of it.)"`
	CommitteeNodes []*dto.ChainNodeStatus `swagger:"desc(Committee nodes and their peering info.)"`
	AccessNodes    []*dto.ChainNodeStatus `swagger:"desc(Access nodes and their peering info.)"`
	CandidateNodes []*dto.ChainNodeStatus `swagger:"desc(Candidate nodes and their peering info.)"`
}
