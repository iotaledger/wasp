// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"errors"
	"time"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/peering/domain"
	"github.com/iotaledger/wasp/packages/peering/group"
	"golang.org/x/xerrors"
)

//
// PeeringNetwork represents a global view of the mocked network.
//
type PeeringNetwork struct {
	nodes     []*peeringNode
	providers []*peeringNetworkProvider
	bufSize   int
	behavior  PeeringNetBehavior
	log       *logger.Logger
}

// NewPeeringNetwork creates new test network, it can then be used to create network nodes.
func NewPeeringNetwork(
	netIDs []string,
	nodeIdentities []*cryptolib.KeyPair,
	bufSize int,
	behavior PeeringNetBehavior,
	log *logger.Logger,
) *PeeringNetwork {
	nodes := make([]*peeringNode, len(netIDs))
	providers := make([]*peeringNetworkProvider, len(netIDs))
	network := PeeringNetwork{
		nodes:     nodes,
		providers: providers,
		bufSize:   bufSize,
		behavior:  behavior,
		log:       log,
	}
	for i := range nodes {
		nodes[i] = newPeeringNode(netIDs[i], nodeIdentities[i], &network)
	}
	for i := range nodes {
		providers[i] = newPeeringNetworkProvider(nodes[i], &network)
	}
	return &network
}

// NetworkProviders returns network providers for each of the nodes in the network.
func (p *PeeringNetwork) NetworkProviders() []peering.NetworkProvider {
	cp := make([]peering.NetworkProvider, len(p.providers))
	for i := range p.providers {
		cp[i] = p.providers[i]
	}
	return cp
}

func (p *PeeringNetwork) nodeByPubKey(nodePubKey *cryptolib.PublicKey) *peeringNode {
	for i := range p.nodes {
		if p.nodes[i].identity.GetPublicKey().Equals(nodePubKey) {
			return p.nodes[i]
		}
	}
	return nil
}

// Close implements the io.Closer interface.
func (p *PeeringNetwork) Close() error {
	for _, n := range p.nodes {
		if err := n.Close(); err != nil {
			return xerrors.Errorf("failed to close test peering node: %w", err)
		}
	}
	p.behavior.Close()
	return nil
}

//
// peeringNode stands for a mock of a node in a fake network.
// It does NOT implement the peering.PeerSender, because the source
// node should be known for the sender.
//
type peeringNode struct {
	netID    string
	identity *cryptolib.KeyPair
	sendCh   chan *peeringMsg
	recvCh   chan *peeringMsg
	recvCbs  []*peeringCb
	network  *PeeringNetwork
	log      *logger.Logger
}

type peeringMsg struct {
	from      *cryptolib.PublicKey
	msg       peering.PeerMessageData
	timestamp int64
}

type peeringCb struct {
	callback  func(recv *peering.PeerMessageIn) // Receive callback.
	destNP    *peeringNetworkProvider           // Destination node.
	peeringID *peering.PeeringID                // Only listen for specific chain msgs.
	receiver  byte
}

func newPeeringNode(netID string, identity *cryptolib.KeyPair, network *PeeringNetwork) *peeringNode {
	sendCh := make(chan *peeringMsg, network.bufSize)
	recvCh := make(chan *peeringMsg, network.bufSize)
	recvCbs := make([]*peeringCb, 0)
	n := peeringNode{
		netID:    netID,
		identity: identity,
		sendCh:   sendCh,
		recvCh:   recvCh,
		recvCbs:  recvCbs,
		network:  network,
		log:      network.log.With("loc", netID),
	}
	network.behavior.AddLink(sendCh, recvCh, identity.GetPublicKey())
	go n.recvLoop()
	return &n
}

func (n *peeringNode) recvLoop() {
	for pm := range n.recvCh {
		msgPeeringID := pm.msg.PeeringID.String()
		for _, cb := range n.recvCbs {
			if cb.peeringID.String() == msgPeeringID && cb.receiver == pm.msg.MsgReceiver {
				cb.callback(&peering.PeerMessageIn{
					PeerMessageData: pm.msg,
					SenderPubKey:    pm.from,
				})
			}
		}
	}
}

func (n *peeringNode) sendMsg(from *cryptolib.PublicKey, msg *peering.PeerMessageData) {
	n.sendCh <- &peeringMsg{
		from: from,
		msg:  *msg,
	}
}

func (n *peeringNode) Close() error {
	close(n.recvCh)
	return nil
}

//
// peeringNetworkProvider to be used in tests as a mock for the peering network.
//
type peeringNetworkProvider struct {
	self    *peeringNode
	network *PeeringNetwork
	senders []*peeringSender // Senders for all the nodes.
}

var _ peering.NetworkProvider = &peeringNetworkProvider{}

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

// Run implements peering.NetworkProvider.
func (p *peeringNetworkProvider) Run(stopCh <-chan struct{}) {
	<-stopCh
}

// Self implements peering.NetworkProvider.
func (p *peeringNetworkProvider) Self() peering.PeerSender {
	return newPeeringSender(p.self, p)
}

// PeerGroup implements peering.NetworkProvider.
func (p *peeringNetworkProvider) PeerGroup(peeringID peering.PeeringID, peerPubKeys []*cryptolib.PublicKey) (peering.GroupProvider, error) {
	peers := make([]peering.PeerSender, len(peerPubKeys))
	for i := range peerPubKeys {
		n := p.network.nodeByPubKey(peerPubKeys[i])
		if n == nil {
			return nil, errors.New("unknown node location")
		}
		peers[i] = p.senders[i]
	}
	return group.NewPeeringGroupProvider(p, peeringID, peers, p.network.log)
}

// PeerDomain creates peering.PeerDomainProvider.
func (p *peeringNetworkProvider) PeerDomain(peeringID peering.PeeringID, peerPubKeys []*cryptolib.PublicKey) (peering.PeerDomainProvider, error) {
	peers := make([]peering.PeerSender, len(peerPubKeys))
	for i := range peerPubKeys {
		n := p.network.nodeByPubKey(peerPubKeys[i])
		if n == nil {
			return nil, errors.New("unknown node pub key")
		}
		peers[i] = p.senders[i]
	}
	return domain.NewPeerDomain(p, peeringID, peers, p.network.log), nil
}

// Attach implements peering.NetworkProvider.
func (p *peeringNetworkProvider) Attach(
	peeringID *peering.PeeringID,
	receiver byte,
	callback func(recv *peering.PeerMessageIn),
) interface{} {
	p.self.recvCbs = append(p.self.recvCbs, &peeringCb{
		callback:  callback,
		destNP:    p,
		peeringID: peeringID,
		receiver:  receiver,
	})
	return nil // We don't care on the attachIDs for now.
}

// Detach implements peering.NetworkProvider.
func (p *peeringNetworkProvider) Detach(attachID interface{}) {
	// Detach is not important in tests.
}

func (p *peeringNetworkProvider) SendMsgByPubKey(peerPubKey *cryptolib.PublicKey, msg *peering.PeerMessageData) {
	s, err := p.PeerByPubKey(peerPubKey)
	if err == nil {
		s.SendMsg(msg)
	}
}

// PeerByNetID implements peering.NetworkProvider.
func (p *peeringNetworkProvider) PeerByNetID(peerNetID string) (peering.PeerSender, error) {
	if s := p.senderByNetID(peerNetID); s != nil {
		return s, nil
	}
	return nil, errors.New("peer not found by NetID")
}

// PeerByNetID implements peering.NetworkProvider.
func (p *peeringNetworkProvider) PeerByPubKey(peerPub *cryptolib.PublicKey) (peering.PeerSender, error) {
	for i := range p.senders {
		if p.senders[i].node.identity.GetPublicKey().Equals(peerPub) {
			return p.senders[i], nil
		}
	}
	return nil, errors.New("peer not found by pubKey")
}

// PeerStatus implements peering.NetworkProvider.
func (p *peeringNetworkProvider) PeerStatus() []peering.PeerStatusProvider {
	peerStatus := make([]peering.PeerStatusProvider, len(p.senders))
	for i := range peerStatus {
		peerStatus[i] = p.senders[i]
	}
	return peerStatus
}

func (p *peeringNetworkProvider) senderByNetID(peerNetID string) *peeringSender {
	for i := range p.senders {
		if p.senders[i].node.netID == peerNetID {
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

var _ peering.PeerSender = &peeringSender{}

func newPeeringSender(node *peeringNode, netProvider *peeringNetworkProvider) *peeringSender {
	return &peeringSender{
		node:        node,
		netProvider: netProvider,
	}
}

// NetID implements peering.PeerSender.
func (p *peeringSender) NetID() string {
	return p.node.netID
}

// PubKey implements peering.PeerSender.
func (p *peeringSender) PubKey() *cryptolib.PublicKey {
	return p.node.identity.GetPublicKey()
}

// Send implements peering.PeerSender.
func (p *peeringSender) SendMsg(msg *peering.PeerMessageData) {
	p.node.sendMsg(p.netProvider.self.identity.GetPublicKey(), msg)
}

// IsAlive implements peering.PeerSender.
func (p *peeringSender) IsAlive() bool {
	return true // Not needed in tests.
}

// Await implements peering.PeerSender.
func (p *peeringSender) Await(timeout time.Duration) error {
	return nil
}

// IsInbound implements peering.PeerStatusProvider.
func (p *peeringSender) IsInbound() bool {
	return true // Not needed in tests.
}

// NumUsers implements peering.PeerStatusProvider.
func (p *peeringSender) NumUsers() int {
	return 0 // Not needed in tests.
}

// Send implements peering.PeerSender.
func (p *peeringSender) Close() {
	// Not needed in tests.
}

// Status implements peering.PeerSender.
func (p *peeringSender) Status() peering.PeerStatusProvider {
	return p
}
