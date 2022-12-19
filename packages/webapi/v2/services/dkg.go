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

func (d *DKGService) GenerateDistributedKey(peerPublicKeys []*cryptolib.PublicKey, threshold uint16, timeoutInMilliseconds uint32) (*models.DKSharesInfo, error) {
	const roundRetry = 1 * time.Second // Retry for Peer <-> Peer communication.
	const stepRetry = 3 * time.Second  // Retry for Initiator -> Peer communication.

	timeout := time.Duration(timeoutInMilliseconds) * time.Millisecond

	dkShare, err := d.dkgNodeProvider().GenerateDistributedKey(peerPublicKeys, threshold, roundRetry, stepRetry, timeout)
	if err != nil {
		return nil, err
	}

	dkShareInfo, err := d.createDKModel(dkShare)
	if err != nil {
		return nil, err
	}

	return dkShareInfo, err
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

	return dkShareInfo, err
}

func (d *DKGService) createDKModel(dkShare tcrypto.DKShare) (*models.DKSharesInfo, error) {
	sharedBinaryPubKey, err := dkShare.DSSSharedPublic().MarshalBinary()
	if err != nil {
		return nil, err
	}

	pubKeyShares := make([][]byte, len(dkShare.DSSPublicShares()))
	for i := range dkShare.DSSPublicShares() {
		publicKeyShare, err := dkShare.DSSPublicShares()[i].MarshalBinary()
		if err != nil {
			return nil, err
		}
		pubKeyShares[i] = publicKeyShare
	}

	nodePubKeys := make([][]byte, len(dkShare.GetNodePubKeys()))
	for i := range dkShare.GetNodePubKeys() {
		nodePubKeys[i] = dkShare.GetNodePubKeys()[i].AsBytes()
	}

	dkShareInfo := models.DKSharesInfo{
		Address:      dkShare.GetAddress().Bech32(parameters.L1().Protocol.Bech32HRP),
		PeerIndex:    dkShare.GetIndex(),
		PeerPubKeys:  nodePubKeys,
		PubKeyShares: pubKeyShares,
		SharedPubKey: sharedBinaryPubKey,
		Threshold:    dkShare.GetT(),
	}

	return &dkShareInfo, nil
}
