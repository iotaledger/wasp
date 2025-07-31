package services

import (
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/dkg"
	"github.com/iotaledger/wasp/v2/packages/peering"
	"github.com/iotaledger/wasp/v2/packages/registry"
	"github.com/iotaledger/wasp/v2/packages/tcrypto"
	"github.com/iotaledger/wasp/v2/packages/webapi/models"
)

const (
	roundRetry = 1 * time.Second // Retry for Peer <-> Peer communication.
	stepRetry  = 3 * time.Second // Retry for Initiator -> Peer communication.
)

type DKGService struct {
	dkShareRegistryProvider registry.DKShareRegistryProvider
	dkgNodeProvider         dkg.NodeProvider
	trustedNetworkManager   peering.TrustedNetworkManager
}

func NewDKGService(dkShareRegistryProvider registry.DKShareRegistryProvider, dkgNodeProvider dkg.NodeProvider, trustedNetworkManager peering.TrustedNetworkManager) *DKGService {
	return &DKGService{
		dkShareRegistryProvider: dkShareRegistryProvider,
		dkgNodeProvider:         dkgNodeProvider,
		trustedNetworkManager:   trustedNetworkManager,
	}
}

func (d *DKGService) GenerateDistributedKey(peerPubKeysOrNames []string, threshold uint16, timeout time.Duration) (*models.DKSharesInfo, error) {
	trustedPeers, err := d.trustedNetworkManager.TrustedPeersByPubKeyOrName(peerPubKeysOrNames)
	if err != nil {
		return nil, err
	}
	peerPubKeys := lo.Map(trustedPeers, func(tp *peering.TrustedPeer, _ int) *cryptolib.PublicKey {
		return tp.PubKey()
	})

	dkShare, err := d.dkgNodeProvider().GenerateDistributedKey(peerPubKeys, threshold, roundRetry, stepRetry, timeout)
	if err != nil {
		return nil, err
	}

	dkShareInfo, err := d.createDKModel(dkShare)
	if err != nil {
		return nil, err
	}

	return dkShareInfo, nil
}

func (d *DKGService) GetShares(sharedAddress *cryptolib.Address) (*models.DKSharesInfo, error) {
	dkShare, err := d.dkShareRegistryProvider.LoadDKShare(sharedAddress)
	if err != nil {
		return nil, err
	}

	dkShareInfo, err := d.createDKModel(dkShare)
	if err != nil {
		return nil, err
	}

	return dkShareInfo, nil
}

func (d *DKGService) createDKModel(dkShare tcrypto.DKShare) (*models.DKSharesInfo, error) {
	publicKey, err := dkShare.DSSSharedPublic().MarshalBinary()
	if err != nil {
		return nil, err
	}

	dssPublicShares := dkShare.DSSPublicShares()
	pubKeySharesHex := make([]string, len(dssPublicShares))
	for i := range dssPublicShares {
		publicKeyShare, err := dssPublicShares[i].MarshalBinary()
		if err != nil {
			return nil, err
		}

		pubKeySharesHex[i] = hexutil.Encode(publicKeyShare)
	}

	peerIdentities := dkShare.GetNodePubKeys()
	peerIdentitiesHex := make([]string, len(peerIdentities))
	for i := range peerIdentities {
		peerIdentitiesHex[i] = peerIdentities[i].String()
	}

	dkShareInfo := &models.DKSharesInfo{
		Address:         dkShare.GetAddress().String(),
		PeerIdentities:  peerIdentitiesHex,
		PeerIndex:       dkShare.GetIndex(),
		PublicKey:       hexutil.Encode(publicKey),
		PublicKeyShares: pubKeySharesHex,
		Threshold:       dkShare.GetT(),
	}

	return dkShareInfo, nil
}
