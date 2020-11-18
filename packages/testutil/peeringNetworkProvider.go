package testutil

import (
	"errors"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/plugins/peering"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/group/edwards25519"
)

//
// PeeringNetwork represents a global view of the mocked network.
//
type PeeringNetwork struct {
	nodes     []*peeringNode
	providers []*peeringNetworkProvider
	bufSize   int
	log       *logger.Logger
}

// NewPeeringNetworkForLocs creates a test network with new keys, etc.
func NewPeeringNetworkForLocs(peerLocs []string, bufSize int, log *logger.Logger) *PeeringNetwork {
	var suite = edwards25519.NewBlakeSHA256Ed25519() //bn256.NewSuite()
	var peerPubs []kyber.Point = make([]kyber.Point, len(peerLocs))
	var peerSecs []kyber.Scalar = make([]kyber.Scalar, len(peerLocs))
	for i := range peerLocs {
		peerSecs[i] = suite.Scalar().Pick(suite.RandomStream())
		peerPubs[i] = suite.Point().Mul(peerSecs[i], nil)
	}
	return NewPeeringNetwork(peerLocs, peerPubs, peerSecs, bufSize, log)
}

// NewPeeringNetwork creates new test network, it can then be used to create network nodes.
func NewPeeringNetwork(
	locations []string,
	pubKeys []kyber.Point,
	secKeys []kyber.Scalar,
	bufSize int,
	log *logger.Logger,
) *PeeringNetwork {
	nodes := make([]*peeringNode, len(locations))
	providers := make([]*peeringNetworkProvider, len(locations))
	var network = PeeringNetwork{
		nodes:     nodes,
		providers: providers,
		bufSize:   bufSize,
		log:       log,
	}
	for i := range nodes {
		nodes[i] = newPeeringNode(locations[i], pubKeys[i], secKeys[i], &network)
	}
	for i := range nodes {
		providers[i] = newPeeringNetworkProvider(nodes[i], &network)
	}
	return &network
}

// NetworkProviders returns network providers for each of the nodes in the network.
func (p *PeeringNetwork) NetworkProviders() []peering.NetworkProvider {
	copy := make([]peering.NetworkProvider, len(p.providers))
	for i := range p.providers {
		copy[i] = p.providers[i]
	}
	return copy
}

func (p *PeeringNetwork) nodeByLocation(nodeLoc string) *peeringNode {
	for i := range p.nodes {
		if p.nodes[i].location == nodeLoc {
			return p.nodes[i]
		}
	}
	return nil
}

//
// peeringNode stands for a mock of a node in a fake network.
// It does NOT implement the peering.PeerSender, because the source
// node should be known for the sender.
//
type peeringNode struct {
	location string
	pubKey   kyber.Point
	secKey   kyber.Scalar
	recvCh   chan peeringMsg
	recvCbs  []*peeringCb
	network  *PeeringNetwork
	log      *logger.Logger
}
type peeringMsg struct {
	from *peeringNode
	msg  peering.PeerMessage
}
type peeringCb struct {
	callback func(from peering.PeerSender, msg *peering.PeerMessage) // Receive callback.
	destNP   *peeringNetworkProvider                                 // Destination node.
	chainID  coretypes.ChainID                                       // Only listen for specific chain msgs.
}

func newPeeringNode(location string, pubKey kyber.Point, secKey kyber.Scalar, network *PeeringNetwork) *peeringNode {
	recvCh := make(chan peeringMsg, network.bufSize)
	recvCbs := make([]*peeringCb, 0)
	node := peeringNode{
		location: location,
		pubKey:   pubKey,
		secKey:   secKey,
		recvCh:   recvCh,
		recvCbs:  recvCbs,
		network:  network,
		log:      network.log.With("loc", location),
	}
	go func() { // Receive loop.
		for {
			var pm peeringMsg = <-recvCh
			node.log.Debugf(
				"received msgType=%v from=%v, chainID=%v",
				pm.msg.MsgType, pm.from.location, pm.msg.ChainID,
			)
			msgChainID := pm.msg.ChainID.String()
			for _, cb := range node.recvCbs {
				if cb.chainID.String() == msgChainID {
					cb.callback(cb.destNP.senderByLocation(pm.from.location), &pm.msg)
				}
			}
		}
	}()
	return &node
}

func (p *peeringNode) sendMsg(from *peeringNode, msg *peering.PeerMessage) {
	p.recvCh <- peeringMsg{
		from: from,
		msg:  *msg,
	}
}

//
// peeringNetworkProvider to be used in tests as a mock for the peering network.
//
type peeringNetworkProvider struct {
	self    *peeringNode
	network *PeeringNetwork
	senders []*peeringSender // Senders for all the nodes.
}

// NewpeeringNetworkProvider initializes new network provider (a local view).
func newPeeringNetworkProvider(self *peeringNode, network *PeeringNetwork) *peeringNetworkProvider {
	senders := make([]*peeringSender, len(network.nodes))
	netProvider := peeringNetworkProvider{
		self:    self,
		network: network,
		senders: senders,
	}
	for i := range network.nodes {
		senders[i] = newPeeringSender(network.nodes[i], &netProvider)
	}
	return &netProvider
}

// Self implements peering.NetworkProvider.
func (p *peeringNetworkProvider) Self() peering.PeerSender {
	return newPeeringSender(p.self, p)
}

// Group implements peering.NetworkProvider.
func (p *peeringNetworkProvider) Group(peerAddrs []string) (peering.GroupProvider, error) {
	nodes := make([]*peeringNode, len(peerAddrs))
	for i := range peerAddrs {
		n := p.network.nodeByLocation(peerAddrs[i])
		if n == nil {
			return nil, errors.New("unknown_node_location")
		}
		nodes[i] = n
	}
	group := newPeeringGroupProvider(p, nodes)
	return group, nil
}

// Attach implements peering.NetworkProvider.
func (p *peeringNetworkProvider) Attach(
	chainID coretypes.ChainID,
	callback func(from peering.PeerSender, msg *peering.PeerMessage),
) {
	p.self.recvCbs = append(p.self.recvCbs, &peeringCb{
		callback: callback,
		destNP:   p,
		chainID:  chainID,
	})
}

// SendByLocation implements peering.NetworkProvider.
func (p *peeringNetworkProvider) SendByLocation(peerLoc string, msg *peering.PeerMessage) {
	if sender := p.senderByLocation(peerLoc); sender != nil {
		sender.SendMsg(msg)
	}
}

func (p *peeringNetworkProvider) senderByLocation(peerLoc string) *peeringSender {
	for i := range p.senders {
		if p.senders[i].node.location == peerLoc {
			return p.senders[i]
		}
	}
	return nil
}

//
// peeringSender represents a local view of a remote node
// and implements the peering.PeerSender interface.
//
type peeringSender struct {
	node        *peeringNode
	netProvider *peeringNetworkProvider
}

func newPeeringSender(node *peeringNode, netProvider *peeringNetworkProvider) *peeringSender {
	return &peeringSender{
		node:        node,
		netProvider: netProvider,
	}
}

// Location implements peering.PeerSender.
func (p *peeringSender) Location() string {
	return p.node.location
}

// PubKey implements peering.PeerSender.
func (p *peeringSender) PubKey() kyber.Point {
	return p.node.pubKey
}

// Send implements peering.PeerSender.
func (p *peeringSender) SendMsg(msg *peering.PeerMessage) {
	p.node.sendMsg(p.netProvider.self, msg)
}

//
// peeringGroupProvider implements peering.GroupProvider
//
type peeringGroupProvider struct {
	netProvider *peeringNetworkProvider
	nodes       []*peeringNode
}

func newPeeringGroupProvider(netProvider *peeringNetworkProvider, nodes []*peeringNode) *peeringGroupProvider {
	return &peeringGroupProvider{
		netProvider: netProvider,
		nodes:       nodes,
	}
}

// PeerIndex implements peering.GroupProvider.
func (p *peeringGroupProvider) PeerIndex(peer peering.PeerSender) (int, error) {
	return p.PeerIndexByPub(peer.PubKey())
}

// PeerIndexByLoc implements peering.GroupProvider.
func (p *peeringGroupProvider) PeerIndexByLoc(peerLoc string) (int, error) {
	for i := range p.nodes {
		if p.nodes[i].location == peerLoc {
			return i, nil
		}
	}
	return -1, errors.New("peer_not_found_by_loc")
}

// PeerIndexByPub implements peering.GroupProvider.
func (p *peeringGroupProvider) PeerIndexByPub(peerPub kyber.Point) (int, error) {
	for i := range p.nodes {
		if p.nodes[i].pubKey.Equal(peerPub) {
			return i, nil
		}
	}
	return -1, errors.New("peer_not_found_by_pub")
}

// Broadcast implements peering.GroupProvider.
func (p *peeringGroupProvider) Broadcast(msg *peering.PeerMessage) {
	for i := range p.nodes {
		p.nodes[i].sendMsg(p.netProvider.self, msg)
	}
}

// SendMsgByIndex implements peering.GroupProvider.
func (p *peeringGroupProvider) SendMsgByIndex(peerIdx int, msg *peering.PeerMessage) {
	p.nodes[peerIdx].sendMsg(p.netProvider.self, msg)
}
