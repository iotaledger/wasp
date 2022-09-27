package models

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/webapi/v2/dto"
)

type CommitteeInfoResponse struct {
	AccessNodes    []*dto.ChainNodeStatus `swagger:"desc(A list of all access nodes and their peering info.)"`
	Active         bool                   `swagger:"desc(Whether or not the chain is active.)"`
	CandidateNodes []*dto.ChainNodeStatus `swagger:"desc(A list of all candidate nodes and their peering info.)"`
	ChainID        string                 `swagger:"desc(ChainID (bech32-encoded).)"`
	CommitteeNodes []*dto.ChainNodeStatus `swagger:"desc(A list of all committee nodes and their peering info.)"`
	StateAddress   string                 `swagger:"desc(State address, if we are part of it.)"`
}

type ContractInfoResponse struct {
	Description string            `swagger:"desc(The description of the contract.)"`
	HName       isc.Hname         `swagger:"desc(The id (HName(name)) of the contract.)"`
	Name        string            `swagger:"desc(The name of the contract.)"`
	ProgramHash hashing.HashValue `swagger:"desc(The hash of the contract.)"`
}

type ContractListResponse []*ContractInfoResponse

type gasFeePolicy struct {
	GasFeeTokenID     string `swagger:"desc(The gas fee token id. Empty if base token.)"`
	GasPerToken       uint64 `swagger:"desc(The amount of gas per token.)"`
	ValidatorFeeShare uint8  `swagger:"desc(The validator fee share.)"`
}

type ChainInfoResponse struct {
	ChainID         string `swagger:"desc(ChainID (bech32-encoded).)"`
	EVMChainID      uint16 `swagger:"desc(The EVM chain ID)"`
	ChainOwnerID    string `swagger:"desc(The chain owner address (bech32-encoded).)"`
	Description     string `swagger:"desc(The description of the chain.)"`
	GasFeePolicy    *gasFeePolicy
	MaxBlobSize     uint32 `swagger:"desc(The maximum contract blob size.)"`
	MaxEventSize    uint16 `swagger:"desc(The maximum event size.)"`                   // TODO: Clarify
	MaxEventsPerReq uint16 `swagger:"desc(The maximum amount of events per request.)"` // TODO: Clarify
}

type OffLedgerRequestBody struct {
	Request string `swagger:"desc(Offledger Request (base64))"`
}

type ChainListResponse []*ChainInfoResponse

func MapChainInfoResponse(chainInfo *dto.ChainInfo, evmChainID uint16) *ChainInfoResponse {
	gasFeeTokenID := ""

	if chainInfo.GasFeePolicy.GasFeeTokenID != nil {
		gasFeeTokenID = chainInfo.GasFeePolicy.GasFeeTokenID.String()
	}

	chainInfoResponse := ChainInfoResponse{
		ChainID:      chainInfo.ChainID.String(),
		EVMChainID:   evmChainID,
		ChainOwnerID: chainInfo.ChainOwnerID.String(),
		Description:  chainInfo.Description,
		GasFeePolicy: &gasFeePolicy{
			GasFeeTokenID:     gasFeeTokenID,
			GasPerToken:       chainInfo.GasFeePolicy.GasPerToken,
			ValidatorFeeShare: chainInfo.GasFeePolicy.ValidatorFeeShare,
		},
		MaxBlobSize:     chainInfo.MaxBlobSize,
		MaxEventSize:    chainInfo.MaxEventSize,
		MaxEventsPerReq: chainInfo.MaxEventsPerReq,
	}

	return &chainInfoResponse
}
