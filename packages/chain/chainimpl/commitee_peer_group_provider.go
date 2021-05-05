package chainimpl

import (
	"github.com/iotaledger/wasp/packages/chain"
)

type committeePeerGroupProvider struct {
	committee chain.Committee
}

var _ chain.PeerGroupProvider = &committeePeerGroupProvider{}

func newCommitteePeerGroupProvider(c chain.Committee) *committeePeerGroupProvider {
	return &committeePeerGroupProvider{committee: c}
}

func newCommitteePeerGroup(c chain.Committee) *chain.PeerGroup {
	return chain.NewPeerGroup(newCommitteePeerGroupProvider(c))
}

func (*committeePeerGroupProvider) Lock()   {}
func (*committeePeerGroupProvider) Unlock() {}

func (p *committeePeerGroupProvider) NumPeers() uint16 {
	if p.committee == nil {
		return 0
	}
	return p.committee.Size()
}

func (p *committeePeerGroupProvider) SendMsg(targetPeerIndex uint16, msgType byte, msgData []byte) error {
	if p.committee == nil {
		return nil
	}
	return p.committee.SendMsg(targetPeerIndex, msgType, msgData)
}

func (p *committeePeerGroupProvider) NumIsAlive(quorum uint16) bool {
	if p.committee == nil {
		return false
	}
	return p.committee.QuorumIsAlive(quorum)
}
