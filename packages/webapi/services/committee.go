package services

import (
	"errors"

	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/webapi/dto"
	"github.com/iotaledger/wasp/packages/webapi/interfaces"
)

var ErrNotInCommittee = errors.New("this node is not in the committee for the chain")

type CommitteeService struct {
	chainsProvider          chains.Provider
	networkProvider         peering.NetworkProvider
	dkShareRegistryProvider registry.DKShareRegistryProvider
}

func NewCommitteeService(chainsProvider chains.Provider, networkProvider peering.NetworkProvider, dkShareRegistryProvider registry.DKShareRegistryProvider) interfaces.CommitteeService {
	return &CommitteeService{
		chainsProvider:          chainsProvider,
		networkProvider:         networkProvider,
		dkShareRegistryProvider: dkShareRegistryProvider,
	}
}

func (c *CommitteeService) GetPublicKey() *cryptolib.PublicKey {
	return c.networkProvider.Self().PubKey()
}

func (c *CommitteeService) GetCommitteeInfo(chainID isc.ChainID) (*dto.ChainNodeInfo, error) {
	chain, err := c.chainsProvider().Get(chainID)
	if err != nil {
		return nil, err
	}

	committeeInfo := chain.GetCommitteeInfo()
	if committeeInfo == nil {
		return nil, ErrNotInCommittee
	}

	dkShare, err := c.dkShareRegistryProvider.LoadDKShare(committeeInfo.Address)
	if err != nil {
		return nil, err
	}

	peeringStatus := peeringStatusIncludeSelf(c.networkProvider)
	candidateNodes := c.getCandidateNodesAccessNodeInfo(chain.GetCandidateNodes())
	if err != nil {
		return nil, err
	}
	chainNodes := chain.GetChainNodes()
	inChainNodes := make(map[cryptolib.PublicKeyKey]bool)

	//
	// Committee nodes.
	committeeNodes := c.getCommitteeNodes(dkShare, peeringStatus, candidateNodes, inChainNodes)

	//
	// Access nodes: accepted as access nodes and not included in the committee.
	accessNodes := c.getAccessNodes(dkShare, chainNodes, peeringStatus, candidateNodes, inChainNodes)

	//
	// Candidate nodes have supplied applications, but are not included
	// in the committee and to the set of the access nodes.
	filteredCandidateNodes := c.getCandidateNodes(peeringStatus, candidateNodes, inChainNodes)
	if err != nil {
		return nil, err
	}

	chainNodeInfo := dto.ChainNodeInfo{
		Address:        committeeInfo.Address,
		AccessNodes:    accessNodes,
		CandidateNodes: filteredCandidateNodes,
		CommitteeNodes: committeeNodes,
	}

	return &chainNodeInfo, nil
}

func (c *CommitteeService) getCommitteeNodes(
	dkShare tcrypto.DKShare,
	peeringStatus map[cryptolib.PublicKeyKey]peering.PeerStatusProvider,
	candidateNodes map[cryptolib.PublicKeyKey]*governance.AccessNodeInfo,
	inChainNodes map[cryptolib.PublicKeyKey]bool,
) []*dto.ChainNodeStatus {
	nodes := make([]*dto.ChainNodeStatus, 0)

	for _, cmtNodePubKey := range dkShare.GetNodePubKeys() {
		nodeStatus := c.makeChainNodeStatus(cmtNodePubKey, peeringStatus, candidateNodes)

		nodes = append(nodes, nodeStatus)
		inChainNodes[cmtNodePubKey.AsKey()] = true
	}

	return nodes
}

func (c *CommitteeService) getAccessNodes(
	dkShare tcrypto.DKShare,
	chainNodes []peering.PeerStatusProvider,
	peeringStatus map[cryptolib.PublicKeyKey]peering.PeerStatusProvider,
	candidateNodes map[cryptolib.PublicKeyKey]*governance.AccessNodeInfo,
	inChainNodes map[cryptolib.PublicKeyKey]bool,
) []*dto.ChainNodeStatus {
	nodes := make([]*dto.ChainNodeStatus, 0)

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
		nodes = append(nodes, c.makeChainNodeStatus(acnPubKey, peeringStatus, candidateNodes))
		inChainNodes[acnPubKey.AsKey()] = true
	}

	return nodes
}

func (c *CommitteeService) getCandidateNodes(peeringStatus map[cryptolib.PublicKeyKey]peering.PeerStatusProvider, candidateNodes map[cryptolib.PublicKeyKey]*governance.AccessNodeInfo, inChainNodes map[cryptolib.PublicKeyKey]bool) []*dto.ChainNodeStatus {
	nodes := make([]*dto.ChainNodeStatus, 0)

	for _, node := range candidateNodes {
		if _, ok := inChainNodes[node.NodePubKey.AsKey()]; ok {
			continue // Only include unused candidates here.
		}

		nodes = append(nodes, c.makeChainNodeStatus(node.NodePubKey, peeringStatus, candidateNodes))
	}

	return nodes
}

func (c *CommitteeService) getCandidateNodesAccessNodeInfo(chainCandidateNodes []*governance.AccessNodeInfo) map[cryptolib.PublicKeyKey]*governance.AccessNodeInfo {
	candidateNodes := make(map[cryptolib.PublicKeyKey]*governance.AccessNodeInfo)
	for _, chainCandidateNode := range chainCandidateNodes {
		candidateNodes[chainCandidateNode.NodePubKey.AsKey()] = chainCandidateNode
	}

	return candidateNodes
}

func (c *CommitteeService) makeChainNodeStatus(
	pubKey *cryptolib.PublicKey,
	peeringStatus map[cryptolib.PublicKeyKey]peering.PeerStatusProvider,
	candidateNodes map[cryptolib.PublicKeyKey]*governance.AccessNodeInfo,
) *dto.ChainNodeStatus {
	cns := dto.ChainNodeStatus{
		Node: dto.PeeringNodeStatus{
			PublicKey: pubKey,
		},
	}

	if n, ok := peeringStatus[pubKey.AsKey()]; ok {
		cns.Node.PeeringURL = n.PeeringURL()
		cns.Node.IsAlive = n.IsAlive()
		cns.Node.NumUsers = n.NumUsers()

		if c.networkProvider.Self().PubKey().Equals(cns.Node.PublicKey) {
			cns.Node.IsTrusted = true
		}
	}

	if n, ok := candidateNodes[pubKey.AsKey()]; ok {
		cns.ForCommittee = n.ForCommittee
		cns.ForAccess = true
		cns.AccessAPI = n.AccessAPI
	}

	return &cns
}

func peeringStatusIncludeSelf(networkProvider peering.NetworkProvider) map[cryptolib.PublicKeyKey]peering.PeerStatusProvider {
	peeringStatus := make(map[cryptolib.PublicKeyKey]peering.PeerStatusProvider)
	for _, n := range networkProvider.PeerStatus() {
		peeringStatus[n.PubKey().AsKey()] = n
	}
	peeringStatus[networkProvider.Self().PubKey().AsKey()] = networkProvider.Self().Status()

	return peeringStatus
}
