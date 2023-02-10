package services

import (
	"github.com/iotaledger/wasp/packages/chains"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/webapi/dto"
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
		PublicKey:  publicKey,
		PeeringURL: p.networkProvider.Self().PeeringURL(),
		IsTrusted:  isTrustedErr == nil,
	}
}

func (p *PeeringService) GetRegisteredPeers() []*dto.PeeringNodeStatus {
	peers := p.networkProvider.PeerStatus()
	peerModels := make([]*dto.PeeringNodeStatus, len(peers))

	for k, v := range peers {
		isTrustedErr := p.trustedNetworkManager.IsTrustedPeer(v.PubKey())

		peerModels[k] = &dto.PeeringNodeStatus{
			Name:       v.Name(),
			PublicKey:  v.PubKey(),
			PeeringURL: v.PeeringURL(),
			IsAlive:    v.IsAlive(),
			NumUsers:   v.NumUsers(),
			IsTrusted:  isTrustedErr == nil,
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
			Name:       v.Name,
			PublicKey:  v.PubKey(),
			PeeringURL: v.PeeringURL,
			IsTrusted:  true,
		}
	}

	return peers, nil
}

func (p *PeeringService) TrustPeer(name string, publicKey *cryptolib.PublicKey, peeringURL string) (*dto.PeeringNodeIdentity, error) {
	identity, err := p.trustedNetworkManager.TrustPeer(name, publicKey, peeringURL)
	if err != nil {
		return nil, err
	}

	mappedIdentity := &dto.PeeringNodeIdentity{
		Name:       name,
		PublicKey:  identity.PubKey(),
		PeeringURL: identity.PeeringURL,
		IsTrusted:  true,
	}

	return mappedIdentity, nil
}

func (p *PeeringService) DistrustPeer(publicKey *cryptolib.PublicKey) (*dto.PeeringNodeIdentity, error) {
	identity, err := p.trustedNetworkManager.DistrustPeer(publicKey)
	if err != nil {
		return nil, err
	}

	mappedIdentity := &dto.PeeringNodeIdentity{
		PublicKey:  identity.PubKey(),
		PeeringURL: identity.PeeringURL,
		IsTrusted:  false,
	}

	return mappedIdentity, nil
}

func (p *PeeringService) IsPeerTrusted(publicKey *cryptolib.PublicKey) error {
	return p.trustedNetworkManager.IsTrustedPeer(publicKey)
}
