package services

import (
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/dkg"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/webapi/v2/models"
)

const (
	roundRetry = 1 * time.Second // Retry for Peer <-> Peer communication.
	stepRetry  = 3 * time.Second // Retry for Initiator -> Peer communication.
)

type DKGService struct {
	dkShareRegistryProvider registry.DKShareRegistryProvider
	dkgNodeProvider         dkg.NodeProvider
}

func NewDKGService(dkShareRegistryProvider registry.DKShareRegistryProvider, dkgNodeProvider dkg.NodeProvider) *DKGService {
	return &DKGService{
		dkShareRegistryProvider: dkShareRegistryProvider,
		dkgNodeProvider:         dkgNodeProvider,
	}
}

func (d *DKGService) GenerateDistributedKey(peerPublicKeys []*cryptolib.PublicKey, threshold uint16, timeout time.Duration) (*models.DKSharesInfo, error) {
	dkShare, err := d.dkgNodeProvider().GenerateDistributedKey(peerPublicKeys, threshold, roundRetry, stepRetry, timeout)
	if err != nil {
		return nil, err
	}

	dkShareInfo, err := d.createDKModel(dkShare)
	if err != nil {
		return nil, err
	}

	return dkShareInfo, nil
}

func (d *DKGService) GetShares(sharedAddress iotago.Address) (*models.DKSharesInfo, error) {
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

		pubKeySharesHex[i] = iotago.EncodeHex(publicKeyShare)
	}

	peerIdentities := dkShare.GetNodePubKeys()
	peerIdentitiesHex := make([]string, len(peerIdentities))
	for i := range peerIdentities {
		peerIdentitiesHex[i] = peerIdentities[i].String()
	}

	dkShareInfo := &models.DKSharesInfo{
		Address:         dkShare.GetAddress().Bech32(parameters.L1().Protocol.Bech32HRP),
		PeerIdentities:  peerIdentitiesHex,
		PeerIndex:       dkShare.GetIndex(),
		PublicKey:       iotago.EncodeHex(publicKey),
		PublicKeyShares: pubKeySharesHex,
		Threshold:       dkShare.GetT(),
	}

	return dkShareInfo, nil
}
