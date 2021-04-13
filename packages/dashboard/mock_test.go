package dashboard

import (
	"fmt"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/processors"
	"github.com/labstack/echo/v4"
	"go.dedis.ch/kyber/v3"
)

// waspServices is a mock implementation of the WaspServices interface
type waspServices struct{}

func (w *waspServices) ConfigDump() map[string]interface{} {
	return map[string]interface{}{
		"foo": "bar",
	}
}

func (w *waspServices) ExploreAddressBaseURL() string {
	return "http://127.0.0.1:8081/explorer/address"
}

func (w *waspServices) NetworkProvider() peering.NetworkProvider {
	return &peeringNetworkProvider{}
}

func (w *waspServices) GetChain(chainID *coretypes.ChainID) chain.Chain {
	return &mockChain{}
}

func (w *waspServices) GetChainRecords() ([]*registry.ChainRecord, error) {
	return []*registry.ChainRecord{
		{
			ChainID: coretypes.RandomChainID(),
			Active:  true,
		},
	}, nil
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
	return []peering.PeerStatusProvider{
		&peeringNode{},
		&peeringNode{},
		&peeringNode{},
	}
}

type peeringNode struct{}

func (p *peeringNode) IsInbound() bool {
	return false
}

func (p *peeringNode) NumUsers() int {
	return 1
}

func (p *peeringNode) NetID() string {
	return "127.0.0.1:4000"
}

func (p *peeringNode) PubKey() kyber.Point {
	panic("not implemented")
}

func (p *peeringNode) SendMsg(msg *peering.PeerMessage) {
	panic("not implemented")
}

func (p *peeringNode) IsAlive() bool {
	return true
}

func (p *peeringNode) Await(timeout time.Duration) error {
	panic("not implemented")
}

// Close releases the reference to the peer, this informs the network
// implementation, that it can disconnect, cleanup resources, etc.
// You need to get new reference to the peer (PeerSender) to use it again.
func (p *peeringNode) Close() {
	panic("not implemented")
}

func (w *waspServices) CallView(chain chain.Chain, hname coretypes.Hname, fname string, params dict.Dict) (dict.Dict, error) {
	chainID := chain.ID()

	if hname == root.Interface.Hname() && fname == root.FuncGetChainInfo {
		ret := dict.New()
		ret.Set(root.VarChainID, codec.EncodeChainID(*chainID))
		ret.Set(root.VarChainOwnerID, codec.EncodeAgentID(coretypes.NewRandomAgentID()))
		ret.Set(root.VarDescription, codec.EncodeString("mock chain"))
		ret.Set(root.VarFeeColor, codec.EncodeColor(ledgerstate.Color{}))
		ret.Set(root.VarDefaultOwnerFee, codec.EncodeInt64(42))
		ret.Set(root.VarDefaultValidatorFee, codec.EncodeInt64(42))

		dst := collections.NewMap(ret, root.VarContractRegistry)
		for i := 0; i < 5; i++ {
			dst.MustSetAt(coretypes.Hname(uint32(i)).Bytes(), root.EncodeContractRecord(&root.ContractRecord{}))
		}
		return ret, nil
	}

	panic(fmt.Sprintf("mock view call not implemented: %s::%s", hname.String(), fname))
}

type mockChain struct{}

func (m *mockChain) ID() *coretypes.ChainID {
	return coretypes.RandomChainID()
}

func (m *mockChain) GetCommitteeInfo() *chain.CommitteeInfo {
	panic("not implemented")
}

func (m *mockChain) ReceiveMessage(_ interface{}) {
	panic("not implemented")
}

func (m *mockChain) Events() chain.ChainEvents {
	panic("not implemented")
}

func (m *mockChain) Processors() *processors.ProcessorCache {
	panic("not implemented")
}

func (m *mockChain) ReceiveTransaction(_ *ledgerstate.Transaction) {
	panic("not implemented")
}

func (m *mockChain) ReceiveInclusionState(_ ledgerstate.TransactionID, _ ledgerstate.InclusionState) {
	panic("not implemented")
}

func (m *mockChain) ReceiveState(stateOutput *ledgerstate.AliasOutput, timestamp time.Time) {
	panic("not implemented")
}

func (m *mockChain) ReceiveOutput(output ledgerstate.Output) {
	panic("not implemented")
}

func (m *mockChain) Dismiss(reason string) {
	panic("not implemented")
}

func (m *mockChain) IsDismissed() bool {
	panic("not implemented")
}

func (m *mockChain) GetRequestProcessingStatus(id coretypes.RequestID) chain.RequestProcessingStatus {
	panic("not implemented")
}

func (m *mockChain) EventRequestProcessed() *events.Event {
	panic("not implemented")
}

func mockDashboard() (*echo.Echo, *Dashboard) {
	e := echo.New()
	d := Init(e, &waspServices{})
	return e, d
}
