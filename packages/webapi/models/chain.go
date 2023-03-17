package models

import (
	"encoding/base64"

	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/packages/webapi/dto"
)

type CommitteeNode struct {
	AccessAPI string                    `json:"accessAPI" swagger:"required"`
	Node      PeeringNodeStatusResponse `json:"node" swagger:"required"`
}

func MapCommitteeNode(status *dto.ChainNodeStatus) CommitteeNode {
	return CommitteeNode{
		AccessAPI: status.AccessAPI,
		Node: PeeringNodeStatusResponse{
			Name:       status.Node.Name,
			IsAlive:    status.Node.IsAlive,
			PeeringURL: status.Node.PeeringURL,
			NumUsers:   status.Node.NumUsers,
			PublicKey:  status.Node.PublicKey.String(),
			IsTrusted:  status.Node.IsTrusted,
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
	AccessNodes    []CommitteeNode `json:"accessNodes" swagger:"desc(A list of all access nodes and their peering info.),required"`
	Active         bool            `json:"active" swagger:"desc(Whether or not the chain is active.),required"`
	CandidateNodes []CommitteeNode `json:"candidateNodes" swagger:"desc(A list of all candidate nodes and their peering info.),required"`
	ChainID        string          `json:"chainId" swagger:"desc(ChainID (Bech32-encoded).),required"`
	CommitteeNodes []CommitteeNode `json:"committeeNodes" swagger:"desc(A list of all committee nodes and their peering info.),required"`
	StateAddress   string          `json:"stateAddress" swagger:"desc(State address, if we are part of it.),required"`
}

type ContractInfoResponse struct {
	Description string `json:"description" swagger:"desc(The description of the contract.),required"`
	HName       string `json:"hName" swagger:"desc(The id (HName as Hex)) of the contract.),required"`
	Name        string `json:"name" swagger:"desc(The name of the contract.),required"`
	ProgramHash string `json:"programHash" swagger:"desc(The hash of the contract. (Hex encoded)),required"`
}

type ChainInfoResponse struct {
	IsActive       bool           `json:"isActive" swagger:"desc(Whether or not the chain is active.),required"`
	ChainID        string         `json:"chainID" swagger:"desc(ChainID (Bech32-encoded).),required"`
	EVMChainID     uint16         `json:"evmChainId" swagger:"desc(The EVM chain ID),required,min(1)"`
	ChainOwnerID   string         `json:"chainOwnerId" swagger:"desc(The chain owner address (Bech32-encoded).),required"`
	GasFeePolicy   *gas.FeePolicy `json:"gasFeePolicy" swagger:"desc(The gas fee policy),required"`
	GasLimits      *gas.Limits    `json:"gasLimits" swagger:"desc(The gas limits),required"`
	CustomMetadata string         `json:"customMetadata" swagger:"desc((base64) Optional extra metadata that is appended to the L1 AliasOutput)"`
}

type StateResponse struct {
	State string `json:"state" swagger:"desc(The state of the requested key (Hex-encoded)),required"`
}

func MapChainInfoResponse(chainInfo *dto.ChainInfo, evmChainID uint16) ChainInfoResponse {
	chainInfoResponse := ChainInfoResponse{
		IsActive:       chainInfo.IsActive,
		ChainID:        chainInfo.ChainID.String(),
		EVMChainID:     evmChainID,
		CustomMetadata: base64.StdEncoding.EncodeToString(chainInfo.CustomMetadata),
	}

	if chainInfo.ChainOwnerID != nil {
		chainInfoResponse.ChainOwnerID = chainInfo.ChainOwnerID.String()
	}

	if chainInfo.GasFeePolicy != nil {
		chainInfoResponse.GasFeePolicy = chainInfo.GasFeePolicy
	}

	if chainInfo.GasLimits != nil {
		chainInfoResponse.GasLimits = chainInfo.GasLimits
	}

	return chainInfoResponse
}

type RequestIDResponse struct {
	RequestID string `json:"requestId" swagger:"desc(The request ID of the given transaction ID. (Hex)),required"`
}

type ChainRecord struct {
	IsActive    bool     `json:"isActive" swagger:"required"`
	AccessNodes []string `json:"accessNodes" swagger:"required"`
}
