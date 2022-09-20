package models

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/webapi/v2/dto"
)

type ChainInfo struct {
	ChainID        string                 `swagger:"desc(ChainID (bech32-encoded))"`
	Active         bool                   `swagger:"desc(Whether or not the chain is active)"`
	StateAddress   string                 `swagger:"desc(State address, if we are part of it.)"`
	CommitteeNodes []*dto.ChainNodeStatus `swagger:"desc(Committee nodes and their peering info.)"`
	AccessNodes    []*dto.ChainNodeStatus `swagger:"desc(Access nodes and their peering info.)"`
	CandidateNodes []*dto.ChainNodeStatus `swagger:"desc(Candidate nodes and their peering info.)"`
}

type ContractInfo struct {
	ProgramHash hashing.HashValue
	Description string
	Name        string
	HName       isc.Hname
}

type ContractsResponse []*ContractInfo
