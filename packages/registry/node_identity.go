package registry

import "github.com/iotaledger/wasp/packages/cryptolib"

type NodeIdentity struct {
	nodeIdentity *cryptolib.KeyPair
}

func NewNodeIdentity(nodeIdentity *cryptolib.KeyPair) *NodeIdentity {
	return &NodeIdentity{
		nodeIdentity: nodeIdentity,
	}
}

func (p *NodeIdentity) NodeIdentity() *cryptolib.KeyPair {
	return p.nodeIdentity
}

func (p *NodeIdentity) NodePublicKey() *cryptolib.PublicKey {
	return p.nodeIdentity.GetPublicKey()
}
