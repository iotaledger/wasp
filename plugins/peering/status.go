package peering

type Status struct {
	MyNetworkId string
	Peers       []*PeerStatus
}

type PeerStatus struct {
	RemoteLocation string
	IsInbound      bool
	IsAlive        bool
	NumUsers       int
}

func GetStatus() *Status {
	return &Status{
		MyNetworkId: MyNetworkId(),
		Peers:       getPeerStatus(),
	}
}

func getPeerStatus() []*PeerStatus {
	r := make([]*PeerStatus, 0)
	iteratePeers(func(peer *Peer) {
		r = append(r, &PeerStatus{
			RemoteLocation: peer.remoteNetid,
			IsInbound:      peer.isInbound(),
			IsAlive:        peer.IsAlive(),
			NumUsers:       peer.NumUsers(),
		})
	})
	return r
}
