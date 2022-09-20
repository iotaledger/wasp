package services

import (
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	models2 "github.com/iotaledger/wasp/packages/webapi/v2/dto"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
)

type NodeService struct {
	logger *logger.Logger

	networkProvider  peering.NetworkProvider
	registryProvider registry.Provider
}

func NewNodeService(logger *logger.Logger, networkProvider peering.NetworkProvider, registryProvider registry.Provider) interfaces.Node {
	return &NodeService{
		logger: logger,

		networkProvider:  networkProvider,
		registryProvider: registryProvider,
	}
}

func (c *NodeService) GetNodeInfo(chain chain.Chain) (*models2.ChainNodeInfo, error) {
	committeeInfo := chain.GetCommitteeInfo()

	dkShare, err := c.registryProvider().LoadDKShare(committeeInfo.Address)
	if err != nil {
		return nil, err
	}

	chainNodes := chain.GetChainNodes()
	peeringStatus := make(map[cryptolib.PublicKeyKey]peering.PeerStatusProvider)

	for _, n := range c.networkProvider.PeerStatus() {
		peeringStatus[n.PubKey().AsKey()] = n
	}

	candidateNodes := make(map[cryptolib.PublicKeyKey]*governance.AccessNodeInfo)

	for _, n := range chain.GetCandidateNodes() {
		pubKey, err := cryptolib.NewPublicKeyFromBytes(n.NodePubKey)
		if err != nil {
			return nil, err
		}
		candidateNodes[pubKey.AsKey()] = n
	}

	inChainNodes := make(map[cryptolib.PublicKeyKey]bool)

	//
	// Committee nodes.
	committeeNodes := getCommitteeNodes(dkShare, peeringStatus, candidateNodes, inChainNodes)

	//
	// Access nodes: accepted as access nodes and not included in the committee.
	accessNodes := getAccessNodes(dkShare, chainNodes, peeringStatus, candidateNodes, inChainNodes)

	//
	// Candidate nodes have supplied applications, but are not included
	// in the committee and to the set of the access nodes.
	filteredCandidateNodes, err := getCandidateNodes(peeringStatus, candidateNodes, inChainNodes)
	if err != nil {
		return nil, err
	}

	chainNodeInfo := models2.ChainNodeInfo{
		AccessNodes:    accessNodes,
		CandidateNodes: filteredCandidateNodes,
		CommitteeNodes: committeeNodes,
	}

	return &chainNodeInfo, nil
}

func getCommitteeNodes(
	dkShare tcrypto.DKShare,
	peeringStatus map[cryptolib.PublicKeyKey]peering.PeerStatusProvider,
	candidateNodes map[cryptolib.PublicKeyKey]*governance.AccessNodeInfo,
	inChainNodes map[cryptolib.PublicKeyKey]bool,
) []*models2.ChainNodeStatus {
	nodes := make([]*models2.ChainNodeStatus, 0)

	for _, cmtNodePubKey := range dkShare.GetNodePubKeys() {
		nodes = append(nodes, makeChainNodeStatus(cmtNodePubKey, peeringStatus, candidateNodes))
		inChainNodes[cmtNodePubKey.AsKey()] = true
	}

	return nodes
}

func getAccessNodes(
	dkShare tcrypto.DKShare,
	chainNodes []peering.PeerStatusProvider,
	peeringStatus map[cryptolib.PublicKeyKey]peering.PeerStatusProvider,
	candidateNodes map[cryptolib.PublicKeyKey]*governance.AccessNodeInfo,
	inChainNodes map[cryptolib.PublicKeyKey]bool,
) []*models2.ChainNodeStatus {
	nodes := make([]*models2.ChainNodeStatus, 0)

	for _, chainNode := range chainNodes {
		acnPubKey := chainNode.PubKey()
		skip := false
		for _, cmtNodePubKey := range dkShare.GetNodePubKeys() {
			if acnPubKey.AsKey() == cmtNodePubKey.AsKey() {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		nodes = append(nodes, makeChainNodeStatus(acnPubKey, peeringStatus, candidateNodes))
		inChainNodes[acnPubKey.AsKey()] = true
	}

	return nodes
}

func getCandidateNodes(
	peeringStatus map[cryptolib.PublicKeyKey]peering.PeerStatusProvider,
	candidateNodes map[cryptolib.PublicKeyKey]*governance.AccessNodeInfo,
	inChainNodes map[cryptolib.PublicKeyKey]bool,
) ([]*models2.ChainNodeStatus, error) {
	nodes := make([]*models2.ChainNodeStatus, 0)

	for _, c := range candidateNodes {
		pubKey, err := cryptolib.NewPublicKeyFromBytes(c.NodePubKey)
		if err != nil {
			return nil, err
		}

		if _, ok := inChainNodes[pubKey.AsKey()]; ok {
			continue // Only include unused candidates here.
		}

		nodes = append(nodes, makeChainNodeStatus(pubKey, peeringStatus, candidateNodes))
	}

	return nodes, nil
}

func makeChainNodeStatus(
	pubKey *cryptolib.PublicKey,
	peeringStatus map[cryptolib.PublicKeyKey]peering.PeerStatusProvider,
	candidateNodes map[cryptolib.PublicKeyKey]*governance.AccessNodeInfo,
) *models2.ChainNodeStatus {
	cns := models2.ChainNodeStatus{
		Node: models2.PeeringNodeStatus{
			PubKey: pubKey.String(),
		},
	}

	if n, ok := peeringStatus[pubKey.AsKey()]; ok {
		cns.Node.NetID = n.NetID()
		cns.Node.IsAlive = n.IsAlive()
		cns.Node.NumUsers = n.NumUsers()
	}

	if n, ok := candidateNodes[pubKey.AsKey()]; ok {
		cns.ForCommittee = n.ForCommittee
		cns.ForAccess = true
		cns.AccessAPI = n.AccessAPI
	}

	return &cns
}
