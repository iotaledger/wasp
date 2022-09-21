package models

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/packages/webapi/v2/dto"
)

type NodeInfoResponse struct {
	AccessNodes    []*dto.ChainNodeStatus `swagger:"desc(Access nodes and their peering info.)"`
	Active         bool                   `swagger:"desc(Whether or not the chain is active)"`
	CandidateNodes []*dto.ChainNodeStatus `swagger:"desc(Candidate nodes and their peering info.)"`
	ChainID        string                 `swagger:"desc(ChainID (bech32-encoded))"`
	CommitteeNodes []*dto.ChainNodeStatus `swagger:"desc(Committee nodes and their peering info.)"`
	StateAddress   string                 `swagger:"desc(State address, if we are part of it.)"`
}

type ContractInfoResponse struct {
	Description string            `swagger:"desc(The description of the contract)"`
	HName       isc.Hname         `swagger:"desc(The id (HName(name)) of the contract)"`
	Name        string            `swagger:"desc(The name of the contract)"`
	ProgramHash hashing.HashValue `swagger:"desc(The hash of the contract)"`
}

type ContractsResponse []*ContractInfoResponse

type ChainInfoResponse struct {
	ChainID         string
	ChainOwnerID    isc.AgentID
	Description     string
	GasFeePolicy    *gas.GasFeePolicy
	MaxBlobSize     uint32
	MaxEventSize    uint16
	MaxEventsPerReq uint16
}

type ChainListResponse []*ChainInfoResponse
