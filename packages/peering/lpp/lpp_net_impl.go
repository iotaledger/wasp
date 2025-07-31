// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// cSpell:words libp2p peerstore multiformats multiaddr shrinkingmap cryptolib rwutil libp2ppeer libp2ptls infof

// Package lpp implements a peering.NetworkProvider based on the libp2p.
//
// The set of known peers is managed in several places:
//   - TrustManager contains a registry of trusted peers and their pub keys.
//     That's the main reference DB for the authentication. It is persistent.
//   - libp2p.Peerstore -- loaded with addresses and public keys based on the
//     TrustManager. It is used by the libp2p for address resolution and
//     protocol negotiation.
//   - In-memory copy of the trust DB in the netImpl struct (maps: peerNy*)
//     with additional runtime data needed for a fast lookup of peers by
//     their libp2p IDs, as well as for authentication etc.
//
// The main identification of a peer is its public key. The address (peeringURL)
// may change over time (because of NAT, and similar reasons).
package lpp

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"fortio.org/safecast"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	libp2ppeer "github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
	"github.com/libp2p/go-libp2p/p2p/transport/tcp"
	"github.com/multiformats/go-multiaddr"

	"github.com/iotaledger/hive.go/ds/shrinkingmap"
	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/hive.go/runtime/event"

	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/peering"
	"github.com/iotaledger/wasp/v2/packages/peering/domain"
	"github.com/iotaledger/wasp/v2/packages/peering/group"
	"github.com/iotaledger/wasp/v2/packages/util/rwutil"
)

const (
	maintenancePeriod = 1 * time.Second

	lppProtocolPeering   = "/iotaledger/wasp/peering/1.0.0"
	lppProtocolHeartbeat = "/iotaledger/wasp/heartbeat/1.0.0"
)

// netImpl implements a peering.NetworkProvider interface.
type netImpl struct {
	myPeeringURL string                                           // peeringURL of this node.
	lppHost      host.Host                                        // The instance of the libp2p to use.
	port         int                                              // Port to use for peering.
	ctx          context.Context                                  // Context for the libp2p
	ctxCancel    context.CancelFunc                               // A way to close the context.
	peers        *shrinkingmap.ShrinkingMap[libp2ppeer.ID, *peer] // By remotePeer.ID()
	peersLock    *sync.RWMutex
	recvEvents   *event.Event1[*peering.PeerMessageIn] // Used to publish events to all attached clients.
	nodeKeyPair  *cryptolib.KeyPair
	trusted      peering.TrustedNetworkManager
	metrics      peering.Metrics
	log          log.Logger
}

var (
	_ peering.NetworkProvider = &netImpl{}
	_ peering.PeerSender      = &netImpl{}
)

// NewNetworkProvider is a constructor for the TCP based
// peering network implementation.
func NewNetworkProvider(
	myPeeringURL string,
	port int,
	nodeKeyPair *cryptolib.KeyPair,
	trusted peering.TrustedNetworkManager,
	metrics peering.Metrics,
	log log.Logger,
) (peering.NetworkProvider, peering.TrustedNetworkManager, error) {
	privKey, err := crypto.UnmarshalEd25519PrivateKey(nodeKeyPair.GetPrivateKey().AsBytes())
	if err != nil {
		return nil, nil, fmt.Errorf("unable to convert the private key: %w", err)
	}
	ctx, ctxCancel := context.WithCancel(context.Background())
	lppHost, err := libp2p.New(
		libp2p.Identity(privKey),
		libp2p.ListenAddrStrings(
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%v", port),
			fmt.Sprintf("/ip6/::1/tcp/%v", port),
		),
		libp2p.Transport(tcp.NewTCPTransport),
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
	)
	if err != nil {
		ctxCancel()
		return nil, nil, fmt.Errorf("failed to construct libp2p host: %w", err)
	}
	n := &netImpl{
		myPeeringURL: myPeeringURL,
		lppHost:      lppHost,
		ctx:          ctx,
		ctxCancel:    ctxCancel,
		port:         port,
		peers:        shrinkingmap.New[libp2ppeer.ID, *peer](),
		peersLock:    &sync.RWMutex{},
		recvEvents:   nil, // Initialized bellow.
		nodeKeyPair:  nodeKeyPair,
		trusted:      trusted,
		metrics:      metrics,
		log:          log,
	}
	n.recvEvents = event.New1[*peering.PeerMessageIn]()
	//
	// Finish initialization of the libp2p node.
	lppHost.SetStreamHandler(lppProtocolPeering, n.lppPeeringProtocolHandler)
	lppHost.SetStreamHandler(lppProtocolHeartbeat, n.lppHeartbeatProtocolHandler)

	if trusted.IsTrustedPeer(n.PubKey()) != nil {
		selfName := "me"
		log.LogInfof("Adding this node as trusted for itself, name=%v, pubKey=%v, peeringURL=%v", selfName, n.PubKey(), n.myPeeringURL)
		if _, err = trusted.TrustPeer(selfName, n.PubKey(), n.myPeeringURL); err != nil {
			ctxCancel()
			return nil, nil, fmt.Errorf("unable to add self to trusted peers: %w", err)
		}
	}

	trustedPeers, err := trusted.TrustedPeers()
	if err != nil {
		ctxCancel()
		return nil, nil, fmt.Errorf("unable to get trusted peers: %w", err)
	}
	for _, trustedPeer := range trustedPeers {
		if err := n.addPeer(trustedPeer); err != nil {
			ctxCancel()
			return nil, nil, fmt.Errorf("unable to setup trusted peer: %w", err)
		}
	}
	return n, n, nil
}

func (n *netImpl) lppAddToPeerStore(trustedPeer *peering.TrustedPeer) (libp2ppeer.ID, error) {
	lppPeerID, lppPeerPub, err := n.lppTrustedPeerID(trustedPeer)
	if err != nil {
		return "", err
	}

	// If the peer's URL starts with `/`, we consider it a multi-address and use as-is.
	// Otherwise we assume it is `host:port` and we build multi-addresses out of that.
	var addrs []multiaddr.Multiaddr
	if strings.HasPrefix(trustedPeer.PeeringURL, "/") {
		addr, err2 := multiaddr.NewMultiaddr(trustedPeer.PeeringURL)
		if err2 != nil {
			return "", fmt.Errorf("failed to parse multiaddr from peeringURL=%v, error: %w", trustedPeer.PeeringURL, err2)
		}
		addrs = []multiaddr.Multiaddr{addr}
	} else {
		addrs2, err2 := n.makeMultiaddr(trustedPeer.PeeringURL)
		if err2 != nil {
			return "", fmt.Errorf("failed to create multiaddr from peeringURL=%v, error: %w", trustedPeer.PeeringURL, err2)
		}
		addrs = addrs2
	}

	n.log.LogInfof("Registering %v as libp2p PeerID=%v with addresses: %+v", trustedPeer.PeeringURL, lppPeerID, addrs)
	n.lppHost.Peerstore().AddAddrs(lppPeerID, addrs, peerstore.PermanentAddrTTL)
	err = n.lppHost.Peerstore().AddPubKey(lppPeerID, lppPeerPub)
	if err != nil {
		return "", fmt.Errorf("failed add PubKey for peeringURL=%v, error: %w", trustedPeer.PeeringURL, err)
	}
	return lppPeerID, nil
}

func (n *netImpl) makeMultiaddr(peeringURL string) (a []multiaddr.Multiaddr, err error) {
	peerHost, peerPort, err := peering.ParsePeeringURL(peeringURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse trusted peer peeringURL=%v, error: %w", peeringURL, err)
	}

	addrPatterns := []string{
		"/%s/%s/udp/%v/quic",
		"/%s/%s/tcp/%v",
	}
	addrs := make([]multiaddr.Multiaddr, 0)
	for i := range addrPatterns {
		var addrType string
		if ip := net.ParseIP(peerHost); ip == nil {
			addrType = "dns"
		} else {
			if ip.To4() != nil {
				addrType = "ip4"
			} else {
				addrType = "ip6"
			}
		}
		addr, err2 := multiaddr.NewMultiaddr(fmt.Sprintf(addrPatterns[i], addrType, peerHost, peerPort))
		if err2 != nil {
			return nil, fmt.Errorf("failed to make libp2p address for peeringURL=%v, error: %w", peeringURL, err2)
		}
		addrs = append(addrs, addr)
	}
	return addrs, nil
}

func (n *netImpl) lppTrustedPeerID(trustedPeer *peering.TrustedPeer) (libp2ppeer.ID, crypto.PubKey, error) {
	lppPeerPub, err := crypto.UnmarshalEd25519PublicKey(trustedPeer.PubKey().AsBytes())
	if err != nil {
		return "", nil, fmt.Errorf("failed to convert pub key: %w", err)
	}
	lppPeerID, err := libp2ppeer.IDFromPublicKey(lppPeerPub)
	if err != nil {
		return "", nil, fmt.Errorf("failed to make libp2p:peer.ID: %w", err)
	}
	return lppPeerID, lppPeerPub, nil
}

// Handles the incoming messages from the network.
func (n *netImpl) lppPeeringProtocolHandler(stream network.Stream) {
	defer stream.Close()
	n.peersLock.RLock()
	remotePeer, exists := n.peers.Get(stream.Conn().RemotePeer())
	n.peersLock.RUnlock()
	if !exists {
		n.log.LogWarnf("Dropping incoming message from unknown peer: %v", stream.Conn().RemotePeer())
		return
	}
	if !remotePeer.trusted {
		n.log.LogWarnf("Dropping incoming message from untrusted peer: %v", stream.Conn().RemotePeer())
		return
	}
	payload, err := readFrame(stream)
	if err != nil {
		n.log.LogWarnf("Failed to read incoming payload from %v, reason=%v", remotePeer.remotePeeringURL, err)
		return
	}
	peerMsg, err := peering.PeerMessageNetFromBytes(payload) // Do not use the signatures, we have TLS.
	if err != nil {
		n.log.LogWarnf("error while decoding a message, reason=%v", err)
		return
	}
	remotePeer.RecvMsg(peerMsg)
}

func (n *netImpl) lppHeartbeatProtocolHandler(stream network.Stream) {
	defer stream.Close()
	n.peersLock.RLock()
	remotePeer, exists := n.peers.Get(stream.Conn().RemotePeer())
	n.peersLock.RUnlock()
	if !exists {
		n.log.LogWarnf("Dropping incoming heartbeat from unknown peer: %v", stream.Conn().RemotePeer())
		return
	}
	payload, err := readFrame(stream)
	if err != nil {
		n.log.LogWarnf("Failed to read incoming heartbeat payload from %v, reason=%v", remotePeer.remotePeeringURL, err)
		return
	}
	if len(payload) != 1 {
		n.log.LogWarnf("Failed to read incoming heartbeat payload from %v, invalid payload size=%v", remotePeer.remotePeeringURL, len(payload))
		return
	}
	remotePeer.noteReceived()
	if payload[0] != 0 {
		n.lppHeartbeatSend(remotePeer, false)
	}
}

func (n *netImpl) lppHeartbeatSend(peer *peer, ackNeeded bool) {
	stream, err := n.lppHost.NewStream(n.ctx, peer.remoteLppID, lppProtocolHeartbeat)
	if err != nil {
		n.log.LogWarnf("Failed to send heartbeat to %v, cannot allocate stream, reason: %v", peer.remotePeeringURL, err)
		return
	}
	defer stream.Close()
	frame := []byte{0}
	if ackNeeded {
		frame[0] = 1
	}
	if err := writeFrame(stream, frame); err != nil {
		n.log.LogWarnf("Failed to send heartbeat to %v, reason: %v", peer.remotePeeringURL, err)
		return
	}
}

func (n *netImpl) addPeer(trustedPeer *peering.TrustedPeer) error {
	//
	// Configure the libp2p.
	lppPeerID, err := n.lppAddToPeerStore(trustedPeer)
	if err != nil {
		return fmt.Errorf("failed to add peer to libp2p peerstore: %w", err)
	}
	//
	// Setup the in-memory lookup maps.
	n.peersLock.Lock()
	defer n.peersLock.Unlock()
	var p *peer
	var exists bool
	if p, exists = n.peers.Get(lppPeerID); exists {
		p.trust(true)                           // It might be distrusted previously.
		p.setPeeringURL(trustedPeer.PeeringURL) // It might be changed.
	} else {
		p = newPeer(trustedPeer.Name, trustedPeer.PeeringURL, trustedPeer.PubKey(), lppPeerID, n)
		n.peers.Set(lppPeerID, p)
		n.metrics.PeerCount(n.peers.Size())
	}
	return nil
}

// delete peer information from the in-memory structures.
// Should be called when the peer is not used anymore by any users.
func (n *netImpl) delPeerWithoutLock(peer *peer) {
	n.lppHost.Peerstore().ClearAddrs(peer.remoteLppID)
	n.peers.Delete(peer.remoteLppID)
	n.metrics.PeerCount(n.peers.Size())
}

// Run starts listening and communicating with the network.
func (n *netImpl) Run(ctx context.Context) {
	queueRecvStopCh := make(chan bool)
	receiveStopCh := make(chan bool)
	maintenanceStopCh := make(chan bool)
	go n.maintenanceLoop(maintenanceStopCh)

	<-ctx.Done()
	n.ctxCancel()
	close(maintenanceStopCh)
	close(receiveStopCh)
	close(queueRecvStopCh)
}

// Self implements peering.NetworkProvider.
func (n *netImpl) Self() peering.PeerSender {
	return n
}

// Group creates peering.GroupProvider.
func (n *netImpl) PeerGroup(peeringID peering.PeeringID, peerPubKeys []*cryptolib.PublicKey) (peering.GroupProvider, error) {
	var err error
	groupPeers := make([]peering.PeerSender, len(peerPubKeys))
	for i := range peerPubKeys {
		if groupPeers[i], err = n.usePeer(peerPubKeys[i]); err != nil {
			return nil, err
		}
	}
	return group.NewPeeringGroupProvider(n, peeringID, groupPeers, n.log)
}

// Domain creates peering.PeerDomainProvider.
func (n *netImpl) PeerDomain(peeringID peering.PeeringID, peerPubKeys []*cryptolib.PublicKey) (peering.PeerDomainProvider, error) {
	peers := make([]peering.PeerSender, 0, len(peerPubKeys))
	for _, peerPubKey := range peerPubKeys {
		if peerPubKey.Equals(n.Self().PubKey()) {
			continue
		}
		p, err := n.usePeer(peerPubKey)
		if err != nil {
			return nil, err
		}
		peers = append(peers, p)
	}
	return domain.NewPeerDomain(n, peeringID, peers, n.log), nil
}

// SendMsgByPubKey sends a message to the specified peer.
func (n *netImpl) SendMsgByPubKey(pubKey *cryptolib.PublicKey, msg *peering.PeerMessageData) {
	peer, err := n.PeerByPubKey(pubKey)
	if err != nil {
		n.log.LogWarnf("SendMsgByPubKey: PubKey %v is not in the network", pubKey.String())
		return
	}
	peer.SendMsg(msg)
	peer.Close()
}

// Attach implements peering.NetworkProvider.
func (n *netImpl) Attach(peeringID *peering.PeeringID, receiver byte, callback func(recv *peering.PeerMessageIn)) context.CancelFunc {
	unhook := n.recvEvents.Hook(func(recv *peering.PeerMessageIn) {
		if *peeringID == recv.PeeringID && receiver == recv.MsgReceiver {
			callback(recv)
		}
	}).Unhook
	return unhook
}

// PeerByPubKey implements peering.NetworkProvider.
// NOTE: For now, only known nodes can be looked up by PubKey.
func (n *netImpl) PeerByPubKey(peerPubKey *cryptolib.PublicKey) (peering.PeerSender, error) {
	return n.usePeer(peerPubKey)
}

// PeerStatus implements peering.NetworkProvider.
func (n *netImpl) PeerStatus() []peering.PeerStatusProvider {
	n.peersLock.RLock()
	defer n.peersLock.RUnlock()

	peerStatus := make([]peering.PeerStatusProvider, 0, n.peers.Size())
	n.peers.ForEach(func(_ libp2ppeer.ID, p *peer) bool {
		peerStatus = append(peerStatus, p)
		return true
	})

	return peerStatus
}

// PeeringURL implements peering.PeerSender for the Self() node.
func (n *netImpl) Name() string {
	return ""
}

// PeeringURL implements peering.PeerSender for the Self() node.
func (n *netImpl) PeeringURL() string {
	return n.myPeeringURL
}

// PubKey implements peering.PeerSender for the Self() node.
func (n *netImpl) PubKey() *cryptolib.PublicKey {
	return n.nodeKeyPair.GetPublicKey()
}

// SendMsg implements peering.PeerSender for the Self() node.
func (n *netImpl) SendMsg(msg *peering.PeerMessageData) {
	// Don't go via the network, if sending a message to self.
	n.triggerRecvEvents(n.Self().PubKey(), &peering.PeerMessageNet{PeerMessageData: msg})
}

func (n *netImpl) triggerRecvEvents(from *cryptolib.PublicKey, msg *peering.PeerMessageNet) {
	n.recvEvents.Trigger(&peering.PeerMessageIn{
		PeerMessageData: msg.PeerMessageData,
		SenderPubKey:    from,
	})
}

// IsAlive implements peering.PeerSender for the Self() node.
func (n *netImpl) IsAlive() bool {
	return true // This node is alive.
}

// NumUsers implements peering.PeerStatusProvider for the Self() node.
func (n *netImpl) NumUsers() int {
	return 1
}

// Await implements peering.PeerSender for the Self() node.
func (n *netImpl) Await(timeout time.Duration) error {
	return nil // This node is alive immediately.
}

// Status implements peering.PeerSender interface for the remote peers.
func (n *netImpl) Status() peering.PeerStatusProvider {
	return n
}

// Close implements peering.PeerSender for the Self() node.
func (n *netImpl) Close() {
	// We will con close the connection of the own node.
}

// IsTrustedPeer implements the peering.TrustedNetworkManager interface.
func (n *netImpl) IsTrustedPeer(pubKey *cryptolib.PublicKey) error {
	return n.trusted.IsTrustedPeer(pubKey)
}

// TrustPeer implements the peering.TrustedNetworkManager interface.
// It delegates everything to other implementation and updates the connections accordingly.
func (n *netImpl) TrustPeer(name string, pubKey *cryptolib.PublicKey, peeringURL string) (*peering.TrustedPeer, error) {
	if err := peering.ValidateTrustedPeerParams(name, pubKey, peeringURL); err != nil {
		return nil, err
	}
	trustedPeer, err := n.trusted.TrustPeer(name, pubKey, peeringURL)
	if err != nil {
		return trustedPeer, err
	}
	return trustedPeer, n.addPeer(trustedPeer)
}

// DistrustPeer implements the peering.TrustedNetworkManager interface.
// It delegates everything to other implementation and updates the connections accordingly.
func (n *netImpl) DistrustPeer(pubKey *cryptolib.PublicKey) (*peering.TrustedPeer, error) {
	n.peersLock.Lock()
	defer n.peersLock.Unlock()

	n.peers.ForEach(func(_ libp2ppeer.ID, p *peer) bool {
		peerPubKey := p.remotePubKey
		if peerPubKey != nil && pubKey.Equals(peerPubKey) {
			p.trust(false)
		}
		return true
	})

	return n.trusted.DistrustPeer(pubKey)
}

// TrustedPeers implements the peering.TrustedNetworkManager interface.
func (n *netImpl) TrustedPeers() ([]*peering.TrustedPeer, error) {
	return n.trusted.TrustedPeers()
}

func (n *netImpl) TrustedPeersByPubKeyOrName(pubKeysOrNames []string) ([]*peering.TrustedPeer, error) {
	return n.trusted.TrustedPeersByPubKeyOrName(pubKeysOrNames)
}

// TrustedPeersListener implements the peering.TrustedNetworkManager interface.
func (n *netImpl) TrustedPeersListener(callback func([]*peering.TrustedPeer)) context.CancelFunc {
	return n.trusted.TrustedPeersListener(callback)
}

func (n *netImpl) usePeer(remotePubKey *cryptolib.PublicKey) (peering.PeerSender, error) {
	if remotePubKey.Equals(n.nodeKeyPair.GetPublicKey()) {
		return n, nil
	}

	n.peersLock.Lock()
	defer n.peersLock.Unlock()

	var foundPeer *peer
	n.peers.ForEach(func(_ libp2ppeer.ID, p *peer) bool {
		if p.remotePubKey.Equals(remotePubKey) {
			p.usePeer()
			foundPeer = p
			return false
		}
		return true
	})

	if foundPeer != nil {
		return foundPeer, nil
	}

	return nil, fmt.Errorf("peer %v is not trusted", remotePubKey)
}

func (n *netImpl) maintenanceLoop(stopCh chan bool) {
	for {
		select {
		case <-time.After(maintenancePeriod):
			n.peersLock.Lock()
			n.peers.ForEach(func(_ libp2ppeer.ID, p *peer) bool {
				p.maintenanceCheck()
				return true
			})
			n.peersLock.Unlock()
		case <-stopCh:
			return
		}
	}
}

// readFrame differs from ReadBytes because it uses ReadFull instead of Read to read the data.
func readFrame(stream network.Stream) ([]byte, error) {
	var msgLenB [4]byte
	if msgLenN, err := io.ReadFull(stream, msgLenB[:]); err != nil || msgLenN != len(msgLenB) {
		if err != nil {
			return nil, fmt.Errorf("failed to read frame len prefix: %w", err)
		}
		if msgLenN != len(msgLenB) {
			return nil, fmt.Errorf("failed to read frame len prefix: not enough bytes read, %v instead of %v", msgLenN, len(msgLenB))
		}
	}
	msgLen := binary.LittleEndian.Uint32(msgLenB[:])
	msgBuf := make([]byte, msgLen)
	if msgBufN, err := io.ReadFull(stream, msgBuf); err != nil || msgBufN != int(msgLen) {
		if err != nil {
			return nil, fmt.Errorf("failed to read frame payload: %w", err)
		}
		if msgBufN != int(msgLen) {
			return nil, fmt.Errorf("failed to read frame payload: not enough bytes read, %v instead of %v", msgBufN, msgLen)
		}
	}
	return msgBuf, nil
}

func writeFrame(stream network.Stream, payload []byte) error {
	ww := rwutil.NewWriter(stream)
	payloadLen, err := safecast.Convert[uint32](len(payload))
	if err != nil {
		return fmt.Errorf("payload length overflow in writeFrame: %w", err)
	}
	ww.WriteUint32(payloadLen)
	if len(payload) != 0 {
		ww.WriteN(payload)
	}
	return ww.Err
}
