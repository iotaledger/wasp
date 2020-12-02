package chains

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coret"
)

type ChainStatus struct {
	ID           *coret.ChainID
	Color        *balance.Color
	Size         uint16
	Quorum       uint16
	OwnPeerIndex uint16
	NumPeers     uint16
	HasQuorum    bool
	PeerStatus   []*chain.PeerStatus
}

func GetStatus(chainID *coret.ChainID) *ChainStatus {
	c := GetChain(*chainID)
	if c == nil {
		return nil
	}
	return &ChainStatus{
		ID:           c.ID(),
		Color:        c.Color(),
		Size:         c.Size(),
		Quorum:       c.Quorum(),
		OwnPeerIndex: c.OwnPeerIndex(),
		NumPeers:     c.NumPeers(),
		HasQuorum:    c.HasQuorum(),
		PeerStatus:   c.PeerStatus(),
	}
}
