package chainimpl

import (
	"github.com/iotaledger/wasp/packages/chain"
)

type commiteePeerGroupProvider struct {
	committee chain.Committee
}

var _ chain.PeerGroupProvider = &commiteePeerGroupProvider{}

func newCommiteePeerGroupProvider(c chain.Committee) *commiteePeerGroupProvider {
	return &commiteePeerGroupProvider{committee: c}
}

func newCommiteePeerGroup(c chain.Committee) *chain.PeerGroup {
	return chain.NewPeerGroup(newCommiteePeerGroupProvider(c))
}

func (*commiteePeerGroupProvider) Lock()   {}
func (*commiteePeerGroupProvider) Unlock() {}

func (p *commiteePeerGroupProvider) NumPeers() uint16 {
	if p.committee == nil {
		return 0
	}
	return p.committee.Size()
}

func (p *commiteePeerGroupProvider) SendMsg(targetPeerIndex uint16, msgType byte, msgData []byte) error {
	if p.committee == nil {
		return nil
	}
	return p.committee.SendMsg(targetPeerIndex, msgType, msgData)
}

func (p *commiteePeerGroupProvider) NumIsAlive(quorum uint16) bool {
	if p.committee == nil {
		return false
	}
	return p.committee.QuorumIsAlive(quorum)
}
