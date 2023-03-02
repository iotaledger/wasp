package services

import (
	"bytes"
	"errors"

	"github.com/iotaledger/hive.go/app/shutdown"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/webapi/interfaces"
)

type NodeService struct {
	chainRecordRegistryProvider registry.ChainRecordRegistryProvider
	nodeOwnerAddresses          []string
	nodeIdentityProvider        registry.NodeIdentityProvider
	shutdownHandler             *shutdown.ShutdownHandler
	trustedNetworkManager       peering.TrustedNetworkManager
}

func NewNodeService(chainRecordRegistryProvider registry.ChainRecordRegistryProvider, nodeOwnerAddresses []string, nodeIdentityProvider registry.NodeIdentityProvider, shutdownHandler *shutdown.ShutdownHandler, trustedNetworkManager peering.TrustedNetworkManager) interfaces.NodeService {
	return &NodeService{
		chainRecordRegistryProvider: chainRecordRegistryProvider,
		nodeOwnerAddresses:          nodeOwnerAddresses,
		nodeIdentityProvider:        nodeIdentityProvider,
		shutdownHandler:             shutdownHandler,
		trustedNetworkManager:       trustedNetworkManager,
	}
}

func (n *NodeService) AddAccessNode(chainID isc.ChainID, peerPubKeyOrName string) error { // TODO: Check the caller for param names.
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

func (n *NodeService) DeleteAccessNode(chainID isc.ChainID, peerPubKeyOrName string) error { // TODO: Check the caller for param names.
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

func (n *NodeService) SetNodeOwnerCertificate(publicKey *cryptolib.PublicKey, ownerAddress iotago.Address) ([]byte, error) {
	nodeIdentity := n.nodeIdentityProvider.NodeIdentity()

	if !bytes.Equal(nodeIdentity.GetPublicKey().AsBytes(), publicKey.AsBytes()) {
		return nil, errors.New("wrong public key")
	}

	ownerAuthorized := false
	for _, nodeOwnerAddressStr := range n.nodeOwnerAddresses {
		_, nodeOwnerAddress, err := iotago.ParseBech32(nodeOwnerAddressStr)
		if err != nil {
			continue
		}
		if bytes.Equal(isc.BytesFromAddress(ownerAddress), isc.BytesFromAddress(nodeOwnerAddress)) {
			ownerAuthorized = true
			break
		}
	}

	if !ownerAuthorized {
		return nil, errors.New("unauthorized request")
	}

	cert := governance.NewNodeOwnershipCertificate(nodeIdentity, ownerAddress)

	return cert.Bytes(), nil
}

func (n *NodeService) ShutdownNode() {
	n.shutdownHandler.SelfShutdown("wasp was shutdown via API", false)
}
