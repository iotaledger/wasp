package services

import (
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/distkeygen"
	"github.com/iotaledger/wasp/v2/packages/peering"
	"github.com/iotaledger/wasp/v2/packages/registry"
	"github.com/iotaledger/wasp/v2/packages/tcrypto"
	"github.com/iotaledger/wasp/v2/packages/webapi/models"
)

const (
	roundRetry = 1 * time.Second // Retry for Peer <-> Peer communication.
	stepRetry  = 3 * time.Second // Retry for Initiator -> Peer communication.
)

type DistKeyGenerationService struct {
	distKeyPartRegistryProvider registry.DistKeyPartRegistryProvider
	distKeyGenNodeProvider      distkeygen.NodeProvider
	trustedNetworkManager       peering.TrustedNetworkManager
}

func NewDistKeyGenerationService(distKeyPartRegistryProvider registry.DistKeyPartRegistryProvider, distKeyGenNodeProvider distkeygen.NodeProvider, trustedNetworkManager peering.TrustedNetworkManager) *DistKeyGenerationService {
	return &DistKeyGenerationService{
		distKeyPartRegistryProvider: distKeyPartRegistryProvider,
		distKeyGenNodeProvider:      distKeyGenNodeProvider,
		trustedNetworkManager:       trustedNetworkManager,
	}
}

func (d *DistKeyGenerationService) GenerateDistributedKey(peerPubKeysOrNames []string, threshold uint16, timeout time.Duration) (*models.DistKeyPartsInfo, error) {
	trustedPeers, err := d.trustedNetworkManager.TrustedPeersByPubKeyOrName(peerPubKeysOrNames)
	if err != nil {
		return nil, err
	}
	peerPubKeys := lo.Map(trustedPeers, func(tp *peering.TrustedPeer, _ int) *cryptolib.PublicKey {
		return tp.PubKey()
	})

	distKeyPart, err := d.distKeyGenNodeProvider().GenerateDistributedKey(peerPubKeys, threshold, roundRetry, stepRetry, timeout)
	if err != nil {
		return nil, err
	}

	distKeyPartInfo, err := d.createDKModel(distKeyPart)
	if err != nil {
		return nil, err
	}

	return distKeyPartInfo, nil
}

func (d *DistKeyGenerationService) GetShares(sharedAddress *cryptolib.Address) (*models.DistKeyPartsInfo, error) {
	distKeyPart, err := d.distKeyPartRegistryProvider.LoadDistKeyPart(sharedAddress)
	if err != nil {
		return nil, err
	}

	distKeyPartInfo, err := d.createDKModel(distKeyPart)
	if err != nil {
		return nil, err
	}

	return distKeyPartInfo, nil
}

func (d *DistKeyGenerationService) createDKModel(distKeyPart tcrypto.DistibutedKeyPart) (*models.DistKeyPartsInfo, error) {
	publicKey, err := distKeyPart.DSSSharedPublic().MarshalBinary()
	if err != nil {
		return nil, err
	}

	dssPublicShares := distKeyPart.DSSPublicShares()
	pubKeySharesHex := make([]string, len(dssPublicShares))
	for i := range dssPublicShares {
		publicKeyShare, err := dssPublicShares[i].MarshalBinary()
		if err != nil {
			return nil, err
		}

		pubKeySharesHex[i] = hexutil.Encode(publicKeyShare)
	}

	peerIdentities := distKeyPart.GetNodePubKeys()
	peerIdentitiesHex := make([]string, len(peerIdentities))
	for i := range peerIdentities {
		peerIdentitiesHex[i] = peerIdentities[i].String()
	}

	distKeyPartInfo := &models.DistKeyPartsInfo{
		Address:         distKeyPart.GetAddress().String(),
		PeerIdentities:  peerIdentitiesHex,
		PeerIndex:       distKeyPart.GetIndex(),
		PublicKey:       hexutil.Encode(publicKey),
		PublicKeyShares: pubKeySharesHex,
		Threshold:       distKeyPart.GetT(),
	}

	return distKeyPartInfo, nil
}
