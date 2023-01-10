package services

import (
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/webapi/v2/dto"
)

type PeeringService struct {
	chainsProvider        chains.Provider
	networkProvider       peering.NetworkProvider
	trustedNetworkManager peering.TrustedNetworkManager
}

func NewPeeringService(chainsProvider chains.Provider, networkProvider peering.NetworkProvider, trustedNetworkManager peering.TrustedNetworkManager) *PeeringService {
	return &PeeringService{
		chainsProvider:        chainsProvider,
		networkProvider:       networkProvider,
		trustedNetworkManager: trustedNetworkManager,
	}
}

func (p *PeeringService) GetIdentity() *dto.PeeringNodeIdentity {
	publicKey := p.networkProvider.Self().PubKey()
	isTrustedErr := p.trustedNetworkManager.IsTrustedPeer(publicKey)

	return &dto.PeeringNodeIdentity{
		PublicKey: publicKey,
		NetID:     p.networkProvider.Self().NetID(),
		IsTrusted: isTrustedErr == nil,
	}
}

func (p *PeeringService) GetRegisteredPeers() []*dto.PeeringNodeStatus {
	peers := p.networkProvider.PeerStatus()
	peerModels := make([]*dto.PeeringNodeStatus, len(peers))

	for k, v := range peers {
		isTrustedErr := p.trustedNetworkManager.IsTrustedPeer(v.PubKey())

		peerModels[k] = &dto.PeeringNodeStatus{
			PublicKey: v.PubKey(),
			NetID:     v.NetID(),
			IsAlive:   v.IsAlive(),
			NumUsers:  v.NumUsers(),
			IsTrusted: isTrustedErr == nil,
		}
	}

	return peerModels
}

func (p *PeeringService) GetTrustedPeers() ([]*dto.PeeringNodeIdentity, error) {
	trustedPeers, err := p.trustedNetworkManager.TrustedPeers()
	if err != nil {
		return nil, err
	}

	peers := make([]*dto.PeeringNodeIdentity, len(trustedPeers))
	for k, v := range trustedPeers {
		peers[k] = &dto.PeeringNodeIdentity{
			PublicKey: v.PubKey(),
			NetID:     v.NetID,
			IsTrusted: true,
		}
	}

	return peers, nil
}

func (p *PeeringService) TrustPeer(publicKey *cryptolib.PublicKey, netID string) (*dto.PeeringNodeIdentity, error) {
	identity, err := p.trustedNetworkManager.TrustPeer(publicKey, netID)
	if err != nil {
		return nil, err
	}

	mappedIdentity := &dto.PeeringNodeIdentity{
		PublicKey: identity.PubKey(),
		NetID:     identity.NetID,
		IsTrusted: true,
	}

	return mappedIdentity, nil
}

func (p *PeeringService) DistrustPeer(publicKey *cryptolib.PublicKey) (*dto.PeeringNodeIdentity, error) {
	identity, err := p.trustedNetworkManager.DistrustPeer(publicKey)
	if err != nil {
		return nil, err
	}

	mappedIdentity := &dto.PeeringNodeIdentity{
		PublicKey: identity.PubKey(),
		NetID:     identity.NetID,
		IsTrusted: false,
	}

	return mappedIdentity, nil
}

func (p *PeeringService) IsPeerTrusted(publicKey *cryptolib.PublicKey) error {
	return p.trustedNetworkManager.IsTrustedPeer(publicKey)
}
