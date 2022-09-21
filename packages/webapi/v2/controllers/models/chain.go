package models

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/packages/webapi/v2/dto"
)

type NodeInfoResponse struct {
	AccessNodes    []*dto.ChainNodeStatus `swagger:"desc(A list of all access nodes and their peering info.)"`
	Active         bool                   `swagger:"desc(Whether or not the chain is active)"`
	CandidateNodes []*dto.ChainNodeStatus `swagger:"desc(A list of all candidate nodes and their peering info.)"`
	ChainID        string                 `swagger:"desc(ChainID (bech32-encoded))"`
	CommitteeNodes []*dto.ChainNodeStatus `swagger:"desc(A list of all committee nodes and their peering info.)"`
	StateAddress   string                 `swagger:"desc(State address, if we are part of it.)"`
}

type ContractInfoResponse struct {
	Description string            `swagger:"desc(The description of the contract)"`
	HName       isc.Hname         `swagger:"desc(The id (HName(name)) of the contract)"`
	Name        string            `swagger:"desc(The name of the contract)"`
	ProgramHash hashing.HashValue `swagger:"desc(The hash of the contract)"`
}

type ContractListResponse []*ContractInfoResponse

type ChainInfoResponse struct {
	ChainID         string `swagger:"desc(ChainID (bech32-encoded))"`
	ChainOwnerID    string `swagger:"desc(The chain owner address (bech32-encoded))"`
	Description     string `swagger:"desc(The description of the chain)"`
	GasFeePolicy    *gas.GasFeePolicy
	MaxBlobSize     uint32 `swagger:"desc(The maximum contract blob size)"`
	MaxEventSize    uint16 `swagger:"desc(The maximum event size)"`                   // TODO: Clarify
	MaxEventsPerReq uint16 `swagger:"desc(The maximum amount of events per request)"` // TODO: Clarify
}

type ChainListResponse []*ChainInfoResponse
