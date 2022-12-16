package services

import (
	"bytes"
	"errors"

	"github.com/iotaledger/hive.go/core/app/pkg/shutdown"
	"github.com/iotaledger/hive.go/core/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/webapi/v2/interfaces"
)

type NodeService struct {
	logger *logger.Logger

	nodeOwnerAddresses   []string
	nodeIdentityProvider registry.NodeIdentityProvider
	shutdownHandler      *shutdown.ShutdownHandler
}

func NewNodeService(log *logger.Logger, nodeOwnerAddresses []string, nodeIdentityProvider registry.NodeIdentityProvider, shutdownHandler *shutdown.ShutdownHandler) interfaces.NodeService {
	return &NodeService{
		logger:               log,
		nodeOwnerAddresses:   nodeOwnerAddresses,
		nodeIdentityProvider: nodeIdentityProvider,
		shutdownHandler:      shutdownHandler,
	}
}

func (n *NodeService) SetNodeOwnerCertificate(nodePubKey []byte, ownerAddress iotago.Address) ([]byte, error) {
	nodeIdentity := n.nodeIdentityProvider.NodeIdentity()

	if !bytes.Equal(nodeIdentity.GetPublicKey().AsBytes(), nodePubKey) {
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
