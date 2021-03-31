package chainimpl

import (
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/util"
	"math/rand"
)

type peerObj struct {
	committee   chain.Committee
	permutation *util.Permutation16
}

var _ chain.Peers = &peerObj{}

func newPeers(c chain.Committee) chain.Peers {
	ret := &peerObj{committee: c}
	if c != nil {
		var rndBytes [32]byte
		rand.Read(rndBytes[:])
		ret.permutation = util.NewPermutation16(c.Size(), rndBytes[:])
	}
	return ret
}

func (p *peerObj) NumPeers() uint16 {
	if p.committee == nil {
		return 0
	}
	return p.committee.Size()
}

func (p *peerObj) SendMsg(targetPeerIndex uint16, msgType byte, msgData []byte) error {
	if p.committee == nil {
		return nil
	}
	return p.committee.SendMsg(targetPeerIndex, msgType, msgData)
}

//
func (p *peerObj) SendToAllUntilFirstError(msgType byte, msgData []byte) uint16 {
	if p.committee == nil {
		return 0
	}
	for i := uint16(0); i < p.committee.Size(); i++ {
		err := p.committee.SendMsg(p.permutation.Next(), msgType, msgData)
		if err != nil {
			return i
		}
	}
	var rndBytes [32]byte
	rand.Read(rndBytes[:])
	p.permutation.Shuffle(rndBytes[:])
	return p.committee.Size()
}

func (p *peerObj) NumIsAlive(quorum uint16) bool {
	if p.committee == nil {
		return false
	}
	return p.committee.QuorumIsAlive(quorum)
}
