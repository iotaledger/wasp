package models

import (
	"net/url"

	"github.com/iotaledger/wasp/v2/packages/vm/gas"
	"github.com/iotaledger/wasp/v2/packages/webapi/dto"
	"github.com/iotaledger/wasp/v2/packages/webapi/routes"
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
	ChainID        string          `json:"chainId" swagger:"desc(ChainID (Hex Address).),required"`
	CommitteeNodes []CommitteeNode `json:"committeeNodes" swagger:"desc(A list of all committee nodes and their peering info.),required"`
	StateAddress   string          `json:"stateAddress" swagger:"desc(State address, if we are part of it.),required"`
}

type ContractInfoResponse struct {
	HName string `json:"hName" swagger:"desc(The id (HName as Hex)) of the contract.),required"`
	Name  string `json:"name" swagger:"desc(The name of the contract.),required"`
}

type PublicChainMetadata struct {
	EVMJsonRPCURL   string `json:"evmJsonRpcURL" swagger:"desc(The EVM json rpc url),required"`
	EVMWebSocketURL string `json:"evmWebSocketURL" swagger:"desc(The EVM websocket url)),required"`

	Name        string `json:"name" swagger:"desc(The name of the chain),required"`
	Description string `json:"description" swagger:"desc(The description of the chain.),required"`
	Website     string `json:"website" swagger:"desc(The official website of the chain.),required"`
}

// ChainInfoResponse includes the metadata standard
type ChainInfoResponse struct {
	IsActive     bool                `json:"isActive" swagger:"desc(Whether or not the chain is active),required"`
	ChainID      string              `json:"chainID" swagger:"desc(ChainID (Hex Address)),required"`
	EVMChainID   uint16              `json:"evmChainId" swagger:"desc(The EVM chain ID),required,min(1)"`
	ChainAdmin   string              `json:"chainAdmin" swagger:"desc(The chain admin address (Hex Address)),required"`
	GasFeePolicy *gas.FeePolicy      `json:"gasFeePolicy" swagger:"desc(The gas fee policy),required"`
	GasLimits    *gas.Limits         `json:"gasLimits" swagger:"desc(The gas limits),required"`
	PublicURL    string              `json:"publicURL" swagger:"desc(The fully qualified public url leading to the chains metadata),required"`
	Metadata     PublicChainMetadata `json:"metadata" swagger:"desc(The metadata of the chain),required"`
}

type StateResponse struct {
	State string `json:"state" swagger:"desc(The state of the requested key (Hex-encoded)),required"`
}

func mapMetadataUrls(response *ChainInfoResponse) {
	if response.PublicURL == "" {
		return
	}

	if response.Metadata.EVMJsonRPCURL == "" {
		response.Metadata.EVMJsonRPCURL, _ = url.JoinPath(response.PublicURL, routes.EVMJsonRPCPathSuffix)
	}

	if response.Metadata.EVMWebSocketURL == "" {
		publicURL, _ := url.Parse(response.PublicURL)

		if publicURL.Scheme == "http" {
			publicURL.Scheme = "ws"
		} else {
			publicURL.Scheme = "wss"
		}

		response.Metadata.EVMWebSocketURL, _ = url.JoinPath(publicURL.String(), routes.EVMJsonWebSocketPathSuffix)
	}
}

func MapChainInfoResponse(chainInfo *dto.ChainInfo, evmChainID uint16) ChainInfoResponse {
	chainInfoResponse := ChainInfoResponse{
		IsActive:   chainInfo.IsActive,
		ChainID:    chainInfo.ChainID.String(),
		EVMChainID: evmChainID,
		PublicURL:  chainInfo.PublicURL,
		Metadata: PublicChainMetadata{
			EVMJsonRPCURL:   chainInfo.Metadata.EVMJsonRPCURL,
			EVMWebSocketURL: chainInfo.Metadata.EVMWebSocketURL,
			Name:            chainInfo.Metadata.Name,
			Description:     chainInfo.Metadata.Description,
			Website:         chainInfo.Metadata.Website,
		},
		GasLimits:    chainInfo.GasLimits,
		GasFeePolicy: chainInfo.GasFeePolicy,
	}

	if chainInfo.ChainAdmin != nil {
		chainInfoResponse.ChainAdmin = chainInfo.ChainAdmin.String()
	}

	mapMetadataUrls(&chainInfoResponse)

	return chainInfoResponse
}

type RequestIDResponse struct {
	RequestID string `json:"requestId" swagger:"desc(The request ID of the given transaction ID. (Hex)),required"`
}

type ChainRecord struct {
	IsActive    bool     `json:"isActive" swagger:"required"`
	AccessNodes []string `json:"accessNodes" swagger:"required"`
}

type RotateChainRequest struct {
	RotateToAddress *string `json:"rotateToAddress" swagger:"desc(The address of the new committee or empty to cancel attempt to rotate)"`
}
