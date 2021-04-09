package dashboard

import (
	"time"

	"github.com/iotaledger/wasp/packages/peering"
	"github.com/labstack/echo/v4"
	"go.dedis.ch/kyber/v3"
)

type waspServices struct{}

func (w *waspServices) ConfigDump() map[string]interface{} {
	return map[string]interface{}{
		"foo": "bar",
	}
}

func (w *waspServices) ExploreAddressBaseURL() string {
	return "http://127.0.0.1:8081/explorer/address"
}

type peeringNetworkProvider struct{}

func (p *peeringNetworkProvider) Run(stopCh <-chan struct{}) {
	panic("not implemented")
}

func (p *peeringNetworkProvider) Self() peering.PeerSender {
	return &peeringNode{}
}

func (p *peeringNetworkProvider) Group(peerAddrs []string) (peering.GroupProvider, error) {
	panic("not implemented")
}

func (p *peeringNetworkProvider) Attach(peeringID *peering.PeeringID, callback func(recv *peering.RecvEvent)) interface{} {
	panic("not implemented")
}

func (p *peeringNetworkProvider) Detach(attachID interface{}) {
	panic("not implemented")
}

func (p *peeringNetworkProvider) PeerByNetID(peerNetID string) (peering.PeerSender, error) {
	panic("not implemented")
}

func (p *peeringNetworkProvider) PeerByPubKey(peerPub kyber.Point) (peering.PeerSender, error) {
	panic("not implemented")
}

func (p *peeringNetworkProvider) PeerStatus() []peering.PeerStatusProvider {
	panic("not implemented")
}

type peeringNode struct{}

func (p *peeringNode) NetID() string {
	return "127.0.0.1:4000"
}

func (p *peeringNode) PubKey() kyber.Point {
	panic("not implemented") // TODO: Implement
}

func (p *peeringNode) SendMsg(msg *peering.PeerMessage) {
	panic("not implemented") // TODO: Implement
}

func (p *peeringNode) IsAlive() bool {
	panic("not implemented") // TODO: Implement
}

func (p *peeringNode) Await(timeout time.Duration) error {
	panic("not implemented") // TODO: Implement
}

// Close releases the reference to the peer, this informs the network
// implementation, that it can disconnect, cleanup resources, etc.
// You need to get new reference to the peer (PeerSender) to use it again.
func (p *peeringNode) Close() {
	panic("not implemented") // TODO: Implement
}

func (w *waspServices) NetworkProvider() peering.NetworkProvider {
	return &peeringNetworkProvider{}
}

func mockDashboard() (*echo.Echo, *Dashboard) {
	e := echo.New()
	d := Init(e, &waspServices{})
	return e, d
}
