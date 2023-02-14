package services

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/core/app/pkg/shutdown"
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

func findPeerByName(tnm peering.TrustedNetworkManager, peerName string) (*peering.TrustedPeer, error) {
	paramIsPubKey := strings.HasPrefix(peerName, "0x")
	var peerPubKey *cryptolib.PublicKey
	var err error
	if paramIsPubKey {
		peerPubKey, err = cryptolib.NewPublicKeyFromString(peerName)
		if err != nil {
			return nil, fmt.Errorf("cannot parse %v as pubKey: %w", peerName, err)
		}
	}

	peers, err := tnm.TrustedPeers()
	if err != nil {
		return nil, errors.New("error getting trusted peers")
	}

	peer, ok := lo.Find(peers, func(p *peering.TrustedPeer) bool {
		if paramIsPubKey {
			return peerPubKey.Equals(p.PubKey())
		}
		return p.Name == peerName
	})
	if !ok {
		return nil, interfaces.ErrPeerNotFound
	}
	return peer, nil
}

func (n *NodeService) AddAccessNode(chainID isc.ChainID, peerName string) error {
	peer, err := findPeerByName(n.trustedNetworkManager, peerName)
	if err != nil {
		return err
	}

	if _, err = n.chainRecordRegistryProvider.UpdateChainRecord(chainID, func(rec *registry.ChainRecord) bool {
		return rec.AddAccessNode(peer.PubKey())
	}); err != nil {
		return errors.New("error saving chain record")
	}

	return nil
}

func (n *NodeService) DeleteAccessNode(chainID isc.ChainID, peerName string) error {
	peer, err := findPeerByName(n.trustedNetworkManager, peerName)
	if err != nil {
		return err
	}
	if _, err := n.chainRecordRegistryProvider.UpdateChainRecord(chainID, func(rec *registry.ChainRecord) bool {
		return rec.RemoveAccessNode(peer.PubKey())
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
