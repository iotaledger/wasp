// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package dkg

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/group/edwards25519"
	rabin_dkg "go.dedis.ch/kyber/v3/share/dkg/rabin"
	"go.dedis.ch/kyber/v3/sign/bdn"
	"go.dedis.ch/kyber/v3/sign/eddsa"
)

type NodeProvider func() *Node

// Node represents a node, that can participate in a DKG procedure.
// It receives commands from the initiator as a dkg.NodeProvider,
// and communicates with other DKG nodes via the peering network.
type Node struct {
	identity     *ed25519.KeyPair                 // Keys of the current node.
	secKey       kyber.Scalar                     // Derived from the identity.
	pubKey       kyber.Point                      // Derived from the identity.
	blsSuite     Suite                            // Cryptography to use for the Pairing based operations.
	edSuite      rabin_dkg.Suite                  // Cryptography to use for the Ed25519 based operations.
	netProvider  peering.NetworkProvider          // Network to communicate through.
	registry     registry.DKShareRegistryProvider // Where to store the generated keys.
	processes    map[string]*proc                 // Only for introspection.
	procLock     *sync.RWMutex                    // To guard access to the process pool.
	initMsgQueue chan *initiatorInitMsgIn         // Incoming events processed async.
	attachID     interface{}                      // Peering attach ID
	log          *logger.Logger
}

// Init creates new node, that can participate in the DKG procedure.
// The node then can run several DKG procedures.
func NewNode(
	identity *ed25519.KeyPair,
	netProvider peering.NetworkProvider,
	reg registry.DKShareRegistryProvider,
	log *logger.Logger,
) (*Node, error) {
	kyberEdDSSA := eddsa.EdDSA{}
	if err := kyberEdDSSA.UnmarshalBinary(identity.PrivateKey.Bytes()); err != nil {
		return nil, err
	}
	n := Node{
		identity:     identity,
		secKey:       kyberEdDSSA.Secret,
		pubKey:       kyberEdDSSA.Public,
		blsSuite:     tcrypto.DefaultSuite(),
		edSuite:      edwards25519.NewBlakeSHA256Ed25519(),
		netProvider:  netProvider,
		registry:     reg,
		processes:    make(map[string]*proc),
		procLock:     &sync.RWMutex{},
		initMsgQueue: make(chan *initiatorInitMsgIn),
		log:          log,
	}
	n.attachID = netProvider.Attach(&initPeeringID, peering.PeerMessageReceiverDkgInit, n.receiveInitMessage)
	go n.recvLoop()
	return &n, nil
}

func (n *Node) receiveInitMessage(peerMsg *peering.PeerMessageIn) {
	if peerMsg.MsgReceiver != peering.PeerMessageReceiverDkgInit {
		panic(fmt.Errorf("DKG init handler does not accept peer messages of other receiver type %v, message type=%v",
			peerMsg.MsgReceiver, peerMsg.MsgType))
	}
	if peerMsg.MsgType != initiatorInitMsgType {
		panic(fmt.Errorf("Wrong type of DKG init message: %v", peerMsg.MsgType))
	}
	msg := &initiatorInitMsg{}
	if err := msg.fromBytes(peerMsg.MsgData); err != nil {
		n.log.Warnf("Dropping unknown message: %v", peerMsg)
		return
	}
	n.initMsgQueue <- &initiatorInitMsgIn{
		initiatorInitMsg: *msg,
		SenderNetID:      peerMsg.SenderNetID,
	}
}

func (n *Node) Close() {
	close(n.initMsgQueue)
	n.netProvider.Detach(n.attachID)
}

// GenerateDistributedKey takes all the required parameters from the node and initiated the DKG procedure.
// This function is executed on the DKG initiator node (a chosen leader for this DKG instance).
//nolint:funlen,gocritic
func (n *Node) GenerateDistributedKey(
	peerNetIDs []string,
	peerPubs []ed25519.PublicKey,
	threshold uint16,
	roundRetry time.Duration, // Retry for Peer <-> Peer communication.
	stepRetry time.Duration, // Retry for Initiator -> Peer communication.
	timeout time.Duration, // Timeout for the entire procedure.
) (*tcrypto.DKShare, error) {
	n.log.Infof("Starting new DKG procedure, initiator=%v, peers=%+v", n.netProvider.Self().NetID(), peerNetIDs)
	var err error
	peerCount := uint16(len(peerNetIDs))
	//
	// Some validationfor the parameters.
	if peerCount < 1 || threshold < 1 || threshold > peerCount {
		return nil, invalidParams(fmt.Errorf("wrong DKG parameters: N = %d, T = %d", peerCount, threshold))
	}
	if threshold < peerCount/2+1 {
		// Quorum t must be larger than half size in order to avoid more than one valid quorum in committee.
		// For the DKG itself it is enough to have t >= 2
		return nil, invalidParams(fmt.Errorf("wrong DKG parameters: for N = %d value T must be at least %d", peerCount, peerCount/2+1))
	}
	//
	// Setup network connections.
	dkgID := peering.RandomPeeringID()
	var netGroup peering.GroupProvider
	if netGroup, err = n.netProvider.PeerGroup(dkgID, peerNetIDs); err != nil {
		return nil, err
	}
	defer netGroup.Close()
	recvCh := make(chan *peering.PeerMessageIn, peerCount*2)
	attachID := n.netProvider.Attach(&dkgID, peering.PeerMessageReceiverDkg, func(recv *peering.PeerMessageIn) {
		recvCh <- recv
	})
	defer n.netProvider.Detach(attachID)
	rTimeout := stepRetry
	gTimeout := timeout
	if peerPubs == nil {
		// Take the public keys from the peering network, if they were not specified.
		peerPubs = make([]ed25519.PublicKey, peerCount)
		for i, n := range netGroup.AllNodes() {
			if err = n.Await(timeout); err != nil {
				return nil, err
			}
			nPub := n.PubKey()
			if nPub == nil {
				return nil, fmt.Errorf("Have no public key for %v", n.NetID())
			}
			peerPubs[i] = *nPub
		}
	}
	//
	// Initialize the peers.
	if err = n.exchangeInitiatorAcks(netGroup, netGroup.AllNodes(), recvCh, rTimeout, gTimeout, rabinStep0Initialize,
		func(peerIdx uint16, peer peering.PeerSender) {
			n.log.Debugf("Initiator sends step=%v command to %v", rabinStep0Initialize, peer.NetID())
			peer.SendMsg(makePeerMessage(initPeeringID, peering.PeerMessageReceiverDkgInit, rabinStep0Initialize, &initiatorInitMsg{
				dkgRef:       dkgID.String(), // It could be some other identifier.
				peeringID:    dkgID,
				peerNetIDs:   peerNetIDs,
				peerPubs:     peerPubs,
				initiatorPub: n.identity.PublicKey,
				threshold:    threshold,
				timeout:      timeout,
				roundRetry:   roundRetry,
			}))
		},
	); err != nil {
		return nil, err
	}
	//
	// Perform the DKG steps, each step in parallel, all steps sequentially.
	// Step numbering (R) is according to <https://github.com/dedis/kyber/blob/master/share/dkg/rabin/dkg.go>.
	if peerCount > 1 {
		if err = n.exchangeInitiatorStep(netGroup, netGroup.AllNodes(), recvCh, rTimeout, gTimeout, dkgID, rabinStep1R21SendDeals); err != nil {
			return nil, err
		}
		if err = n.exchangeInitiatorStep(netGroup, netGroup.AllNodes(), recvCh, rTimeout, gTimeout, dkgID, rabinStep2R22SendResponses); err != nil {
			return nil, err
		}
		if err = n.exchangeInitiatorStep(netGroup, netGroup.AllNodes(), recvCh, rTimeout, gTimeout, dkgID, rabinStep3R23SendJustifications); err != nil {
			return nil, err
		}
		if err = n.exchangeInitiatorStep(netGroup, netGroup.AllNodes(), recvCh, rTimeout, gTimeout, dkgID, rabinStep4R4SendSecretCommits); err != nil {
			return nil, err
		}
		if err = n.exchangeInitiatorStep(netGroup, netGroup.AllNodes(), recvCh, rTimeout, gTimeout, dkgID, rabinStep5R5SendComplaintCommits); err != nil {
			return nil, err
		}
	}
	//
	// Now get the public keys.
	// This also performs the "6-R6-SendReconstructCommits" step implicitly.
	pubShareResponses := map[int]*initiatorPubShareMsg{}
	if err = n.exchangeInitiatorMsgs(netGroup, netGroup.AllNodes(), recvCh, rTimeout, gTimeout, rabinStep6R6SendReconstructCommits,
		func(peerIdx uint16, peer peering.PeerSender) {
			n.log.Debugf("Initiator sends step=%v command to %v", rabinStep6R6SendReconstructCommits, peer.NetID())
			peer.SendMsg(makePeerMessage(dkgID, peering.PeerMessageReceiverDkg, rabinStep6R6SendReconstructCommits, &initiatorStepMsg{}))
		},
		func(recv *peering.PeerMessageGroupIn, initMsg initiatorMsg) (bool, error) {
			switch msg := initMsg.(type) {
			case *initiatorPubShareMsg:
				pubShareResponses[int(recv.SenderIndex)] = msg
				return true, nil
			default:
				n.log.Errorf("unexpected message type instead of initiatorPubShareMsg: %V", msg)
				return false, errors.New("unexpected message type instead of initiatorPubShareMsg")
			}
		},
	); err != nil {
		return nil, err
	}
	sharedAddress := pubShareResponses[0].sharedAddress
	sharedPublic := pubShareResponses[0].sharedPublic
	publicShares := make([]kyber.Point, peerCount)
	for i := range pubShareResponses {
		if !sharedAddress.Equals(pubShareResponses[i].sharedAddress) {
			return nil, fmt.Errorf("nodes generated different addresses")
		}
		if !sharedPublic.Equal(pubShareResponses[i].sharedPublic) {
			return nil, fmt.Errorf("nodes generated different shared public keys")
		}
		publicShares[i] = pubShareResponses[i].publicShare
		{
			var pubShareBytes []byte
			if pubShareBytes, err = pubShareResponses[i].publicShare.MarshalBinary(); err != nil {
				return nil, err
			}
			err = bdn.Verify(
				n.blsSuite,
				pubShareResponses[i].publicShare,
				pubShareBytes,
				pubShareResponses[i].signature,
			)
			if err != nil {
				return nil, err
			}
		}
	}
	n.log.Debugf("Generated SharedAddress=%v, SharedPublic=%v", sharedAddress, sharedPublic)
	//
	// Commit the keys to persistent storage.
	if err = n.exchangeInitiatorAcks(netGroup, netGroup.AllNodes(), recvCh, rTimeout, gTimeout, rabinStep7CommitAndTerminate,
		func(peerIdx uint16, peer peering.PeerSender) {
			n.log.Debugf("Initiator sends step=%v command to %v", rabinStep7CommitAndTerminate, peer.NetID())
			peer.SendMsg(makePeerMessage(dkgID, peering.PeerMessageReceiverDkg, rabinStep7CommitAndTerminate, &initiatorDoneMsg{
				pubShares: publicShares,
			}))
		},
	); err != nil {
		return nil, err
	}
	dkShare := tcrypto.DKShare{
		Address:       sharedAddress,
		N:             peerCount,
		T:             threshold,
		Index:         nil, // Not meaningful in this case.
		SharedPublic:  sharedPublic,
		PublicCommits: nil, // Not meaningful in this case.
		PublicShares:  publicShares,
		PrivateShare:  nil, // Not meaningful in this case.
	}
	return &dkShare, nil
}

// Async recv is needed to avoid locking on the even publisher (Recv vs Attach in proc).
func (n *Node) recvLoop() {
	for recv := range n.initMsgQueue {
		n.onInitMsg(recv)
	}
}

// onInitMsg is a callback to handle the DKG initialization messages.
func (n *Node) onInitMsg(msg *initiatorInitMsgIn) {
	var err error
	var p *proc
	n.procLock.RLock()
	if _, ok := n.processes[msg.dkgRef]; ok {
		// To have idempotence for retries, we need to consider duplicate
		// messages as success, if process is already created.
		n.procLock.RUnlock()
		n.netProvider.SendMsgByNetID(msg.SenderNetID, makePeerMessage(msg.peeringID, peering.PeerMessageReceiverDkg, msg.step, &initiatorStatusMsg{
			error: nil,
		}))
		return
	}
	n.procLock.RUnlock()
	go func() {
		// This part should be executed async, because it accesses the network again, and can
		// be locked because of the naive implementation of `events.Event`. It locks on all the callbacks.
		n.procLock.Lock()
		if p, err = onInitiatorInit(msg.peeringID, &msg.initiatorInitMsg, n); err == nil {
			n.processes[p.dkgRef] = p
		}
		n.procLock.Unlock()
		n.netProvider.SendMsgByNetID(msg.SenderNetID, makePeerMessage(msg.peeringID, peering.PeerMessageReceiverDkg, msg.step, &initiatorStatusMsg{
			error: err,
		}))
	}()
}

// Called by the DKG process on termination.
func (n *Node) dropProcess(p *proc) bool {
	n.procLock.Lock()
	defer n.procLock.Unlock()
	if found := n.processes[p.dkgRef]; found != nil {
		delete(n.processes, p.dkgRef)
		return true
	}
	return false
}

func (n *Node) exchangeInitiatorStep(
	netGroup peering.GroupProvider,
	peers map[uint16]peering.PeerSender,
	recvCh chan *peering.PeerMessageIn,
	retryTimeout time.Duration,
	giveUpTimeout time.Duration,
	dkgID peering.PeeringID,
	step byte,
) error {
	sendCB := func(peerIdx uint16, peer peering.PeerSender) {
		n.log.Debugf("Initiator sends step=%v command to %v", step, peer.NetID())
		peer.SendMsg(makePeerMessage(dkgID, peering.PeerMessageReceiverDkg, step, &initiatorStepMsg{}))
	}
	return n.exchangeInitiatorAcks(netGroup, peers, recvCh, retryTimeout, giveUpTimeout, step, sendCB)
}

func (n *Node) exchangeInitiatorAcks(
	netGroup peering.GroupProvider,
	peers map[uint16]peering.PeerSender,
	recvCh chan *peering.PeerMessageIn,
	retryTimeout time.Duration,
	giveUpTimeout time.Duration,
	step byte,
	sendCB func(peerIdx uint16, peer peering.PeerSender),
) error {
	recvCB := func(recv *peering.PeerMessageGroupIn, msg initiatorMsg) (bool, error) {
		n.log.Debugf("Initiator recv. step=%v response %v from %v", step, msg, recv.SenderNetID)
		return true, nil
	}
	return n.exchangeInitiatorMsgs(netGroup, peers, recvCh, retryTimeout, giveUpTimeout, step, sendCB, recvCB)
}

func (n *Node) exchangeInitiatorMsgs(
	netGroup peering.GroupProvider,
	peers map[uint16]peering.PeerSender,
	recvCh chan *peering.PeerMessageIn,
	retryTimeout time.Duration,
	giveUpTimeout time.Duration,
	step byte,
	sendCB func(peerIdx uint16, peer peering.PeerSender),
	recvCB func(recv *peering.PeerMessageGroupIn, initMsg initiatorMsg) (bool, error),
) error {
	recvInitCB := func(recv *peering.PeerMessageGroupIn) (bool, error) {
		var err error
		var initMsg initiatorMsg
		var isInitMsg bool
		isInitMsg, initMsg, err = readInitiatorMsg(&recv.PeerMessageData, n.blsSuite)
		if !isInitMsg {
			return false, nil
		}
		if err != nil {
			n.log.Warnf("Failed to read message from %v: %v", recv.SenderNetID, recv.PeerMessageData)
			return false, err
		}
		if !initMsg.IsResponse() {
			return false, nil
		}
		if initMsg.Step() != step {
			return false, nil
		}
		if initMsg.Error() != nil {
			return false, initMsg.Error()
		}
		return recvCB(recv, initMsg)
	}
	return netGroup.ExchangeRound(peers, recvCh, retryTimeout, giveUpTimeout, sendCB, recvInitCB)
}
