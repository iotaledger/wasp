package services

import (
	"context"
	"errors"

	"github.com/iotaledger/hive.go/app/shutdown"
	"github.com/iotaledger/wasp/v2/packages/chains"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/parameters"
	"github.com/iotaledger/wasp/v2/packages/peering"
	"github.com/iotaledger/wasp/v2/packages/registry"
	"github.com/iotaledger/wasp/v2/packages/vm/core/governance"
	"github.com/iotaledger/wasp/v2/packages/webapi/interfaces"
)

type NodeService struct {
	chainRecordRegistryProvider registry.ChainRecordRegistryProvider
	nodeIdentityProvider        registry.NodeIdentityProvider
	chainsProvider              chains.Provider
	shutdownHandler             *shutdown.ShutdownHandler
	trustedNetworkManager       peering.TrustedNetworkManager
	l1ParamsFetcher             parameters.L1ParamsFetcher
}

func NewNodeService(
	chainRecordRegistryProvider registry.ChainRecordRegistryProvider,
	nodeIdentityProvider registry.NodeIdentityProvider,
	chainsProvider chains.Provider,
	shutdownHandler *shutdown.ShutdownHandler,
	trustedNetworkManager peering.TrustedNetworkManager,
	l1ParamsFetcher parameters.L1ParamsFetcher,
) interfaces.NodeService {
	return &NodeService{
		chainRecordRegistryProvider: chainRecordRegistryProvider,
		nodeIdentityProvider:        nodeIdentityProvider,
		chainsProvider:              chainsProvider,
		shutdownHandler:             shutdownHandler,
		trustedNetworkManager:       trustedNetworkManager,
		l1ParamsFetcher:             l1ParamsFetcher,
	}
}

func (n *NodeService) AddAccessNode(chainID isc.ChainID, peerPubKeyOrName string) error {
	peers, err := n.trustedNetworkManager.TrustedPeersByPubKeyOrName([]string{peerPubKeyOrName})
	if err != nil {
		return err
	}

	if _, err = n.chainRecordRegistryProvider.UpdateChainRecord(chainID, func(rec *registry.ChainRecord) bool {
		return rec.AddAccessNode(peers[0].PubKey())
	}); err != nil {
		return errors.New("error saving chain record")
	}

	return nil
}

func (n *NodeService) DeleteAccessNode(chainID isc.ChainID, peerPubKeyOrName string) error {
	peers, err := n.trustedNetworkManager.TrustedPeersByPubKeyOrName([]string{peerPubKeyOrName})
	if err != nil {
		return err
	}

	if _, err := n.chainRecordRegistryProvider.UpdateChainRecord(chainID, func(rec *registry.ChainRecord) bool {
		return rec.RemoveAccessNode(peers[0].PubKey())
	}); err != nil {
		return errors.New("error saving chain record")
	}

	return nil
}

func (n *NodeService) NodeOwnerCertificate() []byte {
	nodeIdentity := n.nodeIdentityProvider.NodeIdentity()
	return governance.NewNodeOwnershipCertificate(nodeIdentity, n.chainsProvider().ValidatorAddress())
}

func (n *NodeService) ShutdownNode() {
	n.shutdownHandler.SelfShutdown("wasp was shutdown via API", false)
}

func (n *NodeService) L1Params(ctx context.Context) (*parameters.L1Params, error) {
	return n.l1ParamsFetcher.GetOrFetchLatest(ctx)
}
