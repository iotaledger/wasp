package committees

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/committee"
)

type Status struct {
	Committees []*CommittteeStatus
}

type CommittteeStatus struct {
	Address      *address.Address
	OwnerAddress *address.Address
	Color        *balance.Color
	Size         uint16
	Quorum       uint16
	OwnPeerIndex uint16
	NumPeers     uint16
	HasQuorum    bool
	PeerStatus   []*committee.PeerStatus
}

func GetStatus() *Status {
	return &Status{
		Committees: getCommittees(),
	}
}

func getCommittees() []*CommittteeStatus {
	r := make([]*CommittteeStatus, 0)
	iterateCommittees(func(c committee.Committee) {
		r = append(r, &CommittteeStatus{
			Address:      c.Address(),
			OwnerAddress: c.OwnerAddress(),
			Color:        c.Color(),
			Size:         c.Size(),
			Quorum:       c.Quorum(),
			OwnPeerIndex: c.OwnPeerIndex(),
			NumPeers:     c.NumPeers(),
			HasQuorum:    c.HasQuorum(),
			PeerStatus:   c.PeerStatus(),
		})
	})
	return r
}
