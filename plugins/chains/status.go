package chains

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
)

type ChainStatus struct {
	ID           *coretypes.ChainID
	OwnerAddress *address.Address
	Color        *balance.Color
	Size         uint16
	Quorum       uint16
	OwnPeerIndex uint16
	NumPeers     uint16
	HasQuorum    bool
	PeerStatus   []*chain.PeerStatus
}

func GetStatus(chainID *coretypes.ChainID) *ChainStatus {
	c := GetChain(*chainID)
	if c == nil {
		return nil
	}
	return &ChainStatus{
		ID:           c.ID(),
		OwnerAddress: c.OwnerAddress(),
		Color:        c.Color(),
		Size:         c.Size(),
		Quorum:       c.Quorum(),
		OwnPeerIndex: c.OwnPeerIndex(),
		NumPeers:     c.NumPeers(),
		HasQuorum:    c.HasQuorum(),
		PeerStatus:   c.PeerStatus(),
	}
}
