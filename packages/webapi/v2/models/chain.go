package models

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/webapi/v2/dto"
)

type CommitteeNode struct {
	AccessAPI string
	Node      PeeringNodeStatusResponse
}

func MapCommitteeNode(status *dto.ChainNodeStatus) CommitteeNode {
	return CommitteeNode{
		AccessAPI: status.AccessAPI,
		Node: PeeringNodeStatusResponse{
			IsAlive:   status.Node.IsAlive,
			NetID:     status.Node.NetID,
			NumUsers:  status.Node.NumUsers,
			PublicKey: status.Node.PublicKey.String(),
			IsTrusted: status.Node.IsTrusted,
		},
	}
}

func MapCommitteeNodes(status []*dto.ChainNodeStatus) []CommitteeNode {
	nodes := make([]CommitteeNode, 0)

	for _, node := range status {
		nodes = append(nodes, MapCommitteeNode(node))
	}

	return nodes
}

type CommitteeInfoResponse struct {
	AccessNodes    []CommitteeNode `json:"accessNodes" swagger:"desc(A list of all access nodes and their peering info.)"`
	Active         bool            `json:"active" swagger:"desc(Whether or not the chain is active.)"`
	CandidateNodes []CommitteeNode `json:"candidateNodes" swagger:"desc(A list of all candidate nodes and their peering info.)"`
	ChainID        string          `json:"chainId" swagger:"desc(ChainID (Bech32-encoded).)"`
	CommitteeNodes []CommitteeNode `json:"committeeNodes" swagger:"desc(A list of all committee nodes and their peering info.)"`
	StateAddress   string          `json:"stateAddress" swagger:"desc(State address, if we are part of it.)"`
}

type ContractInfoResponse struct {
	Description string            `json:"description" swagger:"desc(The description of the contract.)"`
	HName       isc.Hname         `json:"hName" swagger:"desc(The id (HName(name)) of the contract.)"`
	Name        string            `json:"name" swagger:"desc(The name of the contract.)"`
	ProgramHash hashing.HashValue `json:"programHash" swagger:"desc(The hash of the contract.)"`
}

type gasFeePolicy struct {
	GasFeeTokenID     string `json:"gasFeeTokenId" swagger:"desc(The gas fee token id. Empty if base token.)"`
	GasPerToken       uint64 `json:"gasPerToken" swagger:"desc(The amount of gas per token.)"`
	ValidatorFeeShare uint8  `json:"validatorFeeShare" swagger:"desc(The validator fee share.)"`
}

type ChainInfoResponse struct {
	IsActive        bool         `json:"isActive" swagger:"desc(Whether or not the chain is active.)"`
	ChainID         string       `json:"chainID" swagger:"desc(ChainID (Bech32-encoded).)"`
	EVMChainID      uint16       `json:"evmChainId" swagger:"desc(The EVM chain ID)"`
	ChainOwnerID    string       `json:"chainOwnerId" swagger:"desc(The chain owner address (Bech32-encoded).)"`
	Description     string       `json:"description" swagger:"desc(The description of the chain.)"`
	GasFeePolicy    gasFeePolicy `json:"gasFeePolicy"`
	MaxBlobSize     uint32       `json:"maxBlobSize" swagger:"desc(The maximum contract blob size.)"`
	MaxEventSize    uint16       `json:"maxEventSize" swagger:"desc(The maximum event size.)"`                      // TODO: Clarify
	MaxEventsPerReq uint16       `json:"maxEventsPerReq" swagger:"desc(The maximum amount of events per request.)"` // TODO: Clarify
}

func MapChainInfoResponse(chainInfo *dto.ChainInfo, evmChainID uint16) ChainInfoResponse {
	gasFeeTokenID := ""

	if chainInfo.GasFeePolicy.GasFeeTokenID != nil {
		gasFeeTokenID = chainInfo.GasFeePolicy.GasFeeTokenID.String()
	}

	chainInfoResponse := ChainInfoResponse{
		IsActive:     chainInfo.IsActive,
		ChainID:      chainInfo.ChainID.String(),
		EVMChainID:   evmChainID,
		ChainOwnerID: chainInfo.ChainOwnerID.String(),
		Description:  chainInfo.Description,
		GasFeePolicy: gasFeePolicy{
			GasFeeTokenID:     gasFeeTokenID,
			GasPerToken:       chainInfo.GasFeePolicy.GasPerToken,
			ValidatorFeeShare: chainInfo.GasFeePolicy.ValidatorFeeShare,
		},
		MaxBlobSize:     chainInfo.MaxBlobSize,
		MaxEventSize:    chainInfo.MaxEventSize,
		MaxEventsPerReq: chainInfo.MaxEventsPerReq,
	}

	return chainInfoResponse
}

type RequestIDResponse struct {
	RequestID string `json:"requestId" swagger:"desc(The request ID of the given transaction ID.)"`
}
