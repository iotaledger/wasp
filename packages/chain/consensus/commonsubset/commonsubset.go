// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package commonsubset wraps the ACS (Asynchronous Common Subset) part from
// the HoneyBadgerBFT with the Peering for communication layer. The main objective
// here is to ensure eventual delivery of all the messages, as that is the assumption
// of the ACS algorithm.
//
// To ensure the eventual delivery of messages, we are resending them until they
// are acknowledged by a peer. The acknowledgements are carried along with other
// messages to decrease communication overhead.
//
// We are using a forked version of https://github.com/anthdm/hbbft. The fork was
// made to gave misc mistakes fixed that made apparent in the case of out-of-order
// message delivery.
//
// The main references:
//
//  - https://eprint.iacr.org/2016/199.pdf
//    The original HBBFT paper.
//
//  - https://dl.acm.org/doi/10.1145/2611462.2611468
//    Here the BBA part is originally presented.
//    Most of the mistakes in the lib was in this part.
//
//  - https://dl.acm.org/doi/10.1145/197917.198088
//    The definition of ACS is by Ben-Or. At least looks so.
//

//nolint:dupl // TODO there is a bunch of duplicated code in this file, should be refactored to reusable funcs
package commonsubset

import (
	"bytes"
	"encoding/binary"
	"io"
	"time"

	"github.com/anthdm/hbbft"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/chain/consensus/commoncoin"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
)

const (
	resendPeriod = 500 * time.Millisecond
)

// region CommonSubset /////////////////////////////////////////////////////////

// CommonSubject is responsible for executing a single instance of agreement.
// The main task for this object is to manage all the redeliveries and acknowledgements.
type CommonSubset struct {
	impl *hbbft.ACS // The ACS implementation.

	batchCounter   uint32               // Used to assign unique numbers to the messages.
	missingAcks    map[uint32]time.Time // Not yet acknowledged batches sent by us and the time of the last resend.
	pendingAcks    [][]uint32           // pendingAcks[PeerID] -- acks to be sent to the peer.
	recvMsgBatches []map[uint32]uint32  // [PeerId]ReceivedMbID -> AckedInMb (0 -- unacked). It remains 0, if acked in non-data message.
	sentMsgBatches map[uint32]*msgBatch // BatchID -> MsgBatch

	sessionID          uint64 // Unique identifier for this consensus instance. Used to route messages.
	stateIndex         uint32 // Sequence number of the CS transaction we are agreeing on.
	committeePeerGroup peering.GroupProvider
	netOwnIndex        uint16 // Our index in the group.

	inputCh  chan []byte            // For our input to the consensus.
	recvCh   chan *msgBatch         // For incoming messages.
	closeCh  chan bool              // To implement closing of the object.
	outputCh chan map[uint16][]byte // The caller will receive its result via this channel.
	done     bool                   // Indicates, if the decision is already made.
	log      *logger.Logger         // Logger, of course.
}

func NewCommonSubset(
	sessionID uint64,
	stateIndex uint32,
	committeePeerGroup peering.GroupProvider,
	dkShare *tcrypto.DKShare,
	allRandom bool, // Set to true to have real CC rounds for each epoch. That's for testing mostly.
	outputCh chan map[uint16][]byte,
	log *logger.Logger,
) (*CommonSubset, error) {
	ownIndex := committeePeerGroup.SelfIndex()
	allNodes := committeePeerGroup.AllNodes()
	nodeCount := len(allNodes)
	nodes := make([]uint64, nodeCount)
	nodePos := 0
	for ni := range allNodes {
		nodes[nodePos] = uint64(ni)
		nodePos++
	}
	if outputCh == nil {
		outputCh = make(chan map[uint16][]byte, 1)
	}
	var salt [8]byte
	binary.BigEndian.PutUint64(salt[:], sessionID)
	acsCfg := hbbft.Config{
		N:          nodeCount,
		F:          nodeCount - int(dkShare.T),
		ID:         uint64(ownIndex),
		Nodes:      nodes,
		BatchSize:  0, // Unused in ACS.
		CommonCoin: commoncoin.NewBlsCommonCoin(dkShare, salt[:], allRandom),
	}
	cs := CommonSubset{
		impl:               hbbft.NewACS(acsCfg),
		batchCounter:       0,
		missingAcks:        make(map[uint32]time.Time),
		pendingAcks:        make([][]uint32, nodeCount),
		recvMsgBatches:     make([]map[uint32]uint32, nodeCount),
		sentMsgBatches:     make(map[uint32]*msgBatch),
		sessionID:          sessionID,
		stateIndex:         stateIndex,
		committeePeerGroup: committeePeerGroup,
		netOwnIndex:        ownIndex,
		inputCh:            make(chan []byte, 1),
		recvCh:             make(chan *msgBatch, 1),
		closeCh:            make(chan bool),
		outputCh:           outputCh,
		log:                log,
	}
	for i := range cs.recvMsgBatches {
		cs.recvMsgBatches[i] = make(map[uint32]uint32)
	}
	for i := range cs.pendingAcks {
		cs.pendingAcks[i] = make([]uint32, 0)
	}
	go cs.run()
	return &cs, nil
}

// OutputCh returns the output channel, so that the called don't need to track it.
func (cs *CommonSubset) OutputCh() chan map[uint16][]byte {
	return cs.outputCh
}

// Input accepts the current node's proposal for the consensus.
func (cs *CommonSubset) Input(input []byte) {
	cs.inputCh <- input
}

// HandleMsgBatch accepts a parsed msgBatch as an input from other node.
// This function is used in the CommonSubsetCoordinator to avoid parsing
// the received message multiple times.
func (cs *CommonSubset) HandleMsgBatch(mb *msgBatch) {
	defer func() {
		if err := recover(); err != nil {
			// Just to avoid panics on writing to a closed channel.
			// This can happen on the ACS termination.
			cs.log.Warnf("Dropping msgBatch reason=%v", err)
		}
	}()
	cs.recvCh <- mb
}

func (cs *CommonSubset) Close() {
	cs.impl.Stop()
	close(cs.closeCh)
	close(cs.recvCh)
	close(cs.inputCh)
}

func (cs *CommonSubset) run() {
	retry := time.After(resendPeriod)
	for {
		select {
		case <-retry:
			cs.timeTick()
			if !cs.done || len(cs.pendingAcks) != 0 {
				// The condition for stopping the resend is a bit tricky, because it is
				// not enough for this node to complete with the decision. This node
				// must help others to decide as well.
				retry = time.After(resendPeriod)
			}
		case input, ok := <-cs.inputCh:
			if !ok {
				return
			}
			if err := cs.handleInput(input); err != nil {
				cs.log.Errorf("Failed to handle input, reason: %v", err)
			}
		case mb, ok := <-cs.recvCh:
			if !ok {
				return
			}
			cs.handleMsgBatch(mb)
		case <-cs.closeCh:
			return
		}
	}
}

func (cs *CommonSubset) timeTick() {
	if cs.log.Desugar().Core().Enabled(logger.LevelDebug) { // Debug info for the ACS.
		rbcInstances, bbaInstances, rbcResults, bbaResults, msgQueue := cs.impl.DebugInfo()
		cs.log.Debugf("ACS::timeTick[%v], sessionId=%v, done=%v queue.len=%v, bba.res=%v", cs.impl.ID, cs.sessionID, cs.done, len(msgQueue), bbaResults)
		for i, bba := range bbaInstances {
			if !bba.Done() {
				cs.log.Debugf("ACS::timeTick[%v], sessionId=%v, impl.bba[%v]=%+v", cs.impl.ID, cs.sessionID, i, bba)
			}
		}
		for i, rbc := range rbcInstances {
			if _, ok := rbcResults[i]; !ok {
				cs.log.Debugf("ACS::timeTick[%v], sessionId=%v, impl.rbc[%v]=%+v", cs.impl.ID, cs.sessionID, i, rbc)
			}
		}
	}

	now := time.Now()
	resentBefore := now.Add(resendPeriod * (-2))
	for missingAck, lastSentTime := range cs.missingAcks {
		if lastSentTime.Before(resentBefore) {
			if mb, ok := cs.sentMsgBatches[missingAck]; ok {
				cs.send(mb)
				cs.missingAcks[missingAck] = now
			} else {
				// Batch is already cleaned up, so we don't need the ack.
				delete(cs.missingAcks, missingAck)
			}
		}
	}
}

func (cs *CommonSubset) handleInput(input []byte) error {
	var err error
	if err = cs.impl.InputValue(input); err != nil {
		return xerrors.Errorf("Failed to process ACS.InputValue: %w", err)
	}
	cs.sendPendingMessages()
	return nil
}

func (cs *CommonSubset) handleMsgBatch(recvBatch *msgBatch) {
	//
	// Cleanup all the acknowledged messages from the missing acks list.
	for _, receivedAck := range recvBatch.acks {
		delete(cs.missingAcks, receivedAck)
	}
	//
	// Resend the old responses, if ack was not received
	// and the message is already processed.
	if ackIn, ok := cs.recvMsgBatches[recvBatch.src][recvBatch.id]; ok {
		if ackIn == 0 {
			// Make a fake message just to acknowledge the message, because sender has
			// retried it already and we haven't sent the ack yet. We will not expect
			// an acknowledgement for this ack.
			cs.send(cs.makeAckOnlyBatch(recvBatch.src, recvBatch.id))
		} else {
			cs.send(cs.sentMsgBatches[ackIn])
		}
		return
	}
	//
	// If we have completed with the decision, we just acknowledging other's messages.
	// That is needed to help other members to decide on termination.
	if cs.done {
		if recvBatch.NeedsAck() {
			cs.send(cs.makeAckOnlyBatch(recvBatch.src, recvBatch.id))
		}
		return
	}
	if recvBatch.NeedsAck() {
		// Batches with id=0 are for acks only, they contain no data,
		// therefore we will not track them, just respond on the fly.
		// Otherwise we store the batch ID to be acknowledged with the
		// next outgoing message batch to that peer.
		cs.recvMsgBatches[recvBatch.src][recvBatch.id] = 0 // Received, not acknowledged yet.
		cs.pendingAcks[recvBatch.src] = append(cs.pendingAcks[recvBatch.src], recvBatch.id)
	}
	//
	// Process the messages.
	for _, m := range recvBatch.msgs {
		if err := cs.impl.HandleMessage(uint64(recvBatch.src), m); err != nil {
			cs.log.Errorf("Failed to handle message: %v, message=%+v", err, m)
		}
	}
	//
	// Send the outgoing messages.
	cs.sendPendingMessages()
	//
	// Check, maybe we are done.
	if cs.impl.Done() {
		var output map[uint64][]byte = cs.impl.Output()
		out16 := make(map[uint16][]byte)
		for index, share := range output {
			out16[uint16(index)] = share
		}
		cs.done = true
		cs.outputCh <- out16
	}
}

func (cs *CommonSubset) sendPendingMessages() {
	var outBatches []*msgBatch
	var err error
	if outBatches, err = cs.makeBatches(cs.impl.Messages()); err != nil {
		cs.log.Errorf("Failed to make out batch: %v", err)
	}
	now := time.Now()
	for _, b := range outBatches {
		b.acks = cs.pendingAcks[b.dst]
		cs.pendingAcks[b.dst] = make([]uint32, 0)
		for i := range b.acks {
			// Update the reverse index, i.e. for each message,
			// specify message which have acknowledged it.
			cs.recvMsgBatches[b.dst][b.acks[i]] = b.id
		}
		cs.sentMsgBatches[b.id] = b
		cs.missingAcks[b.id] = now
		cs.send(b)
	}
}

func (cs *CommonSubset) makeBatches(msgs []hbbft.MessageTuple) ([]*msgBatch, error) {
	batchMsgs := make([][]*hbbft.ACSMessage, cs.impl.N)
	for i := range batchMsgs {
		batchMsgs[i] = make([]*hbbft.ACSMessage, 0)
	}
	for _, m := range msgs {
		if acsMsg, ok := m.Payload.(*hbbft.ACSMessage); ok {
			batchMsgs[m.To] = append(batchMsgs[m.To], acsMsg)
		} else {
			return nil, xerrors.Errorf("unexpected message payload type: %T", m.Payload)
		}
	}
	batches := make([]*msgBatch, 0)
	for i := range batchMsgs {
		if len(batchMsgs[i]) == 0 {
			continue
		}
		cs.batchCounter++
		batches = append(batches, &msgBatch{
			sessionID:  cs.sessionID,
			stateIndex: cs.stateIndex,
			id:         cs.batchCounter,
			src:        uint16(cs.impl.ID),
			dst:        uint16(i),
			msgs:       batchMsgs[i],
			acks:       make([]uint32, 0), // Filled later.
		})
	}
	return batches, nil
}

func (cs *CommonSubset) makeAckOnlyBatch(peerID uint16, ackMB uint32) *msgBatch {
	acks := []uint32{}
	if ackMB != 0 {
		acks = append(acks, ackMB)
	}
	if len(cs.pendingAcks[peerID]) != 0 {
		acks = append(acks, cs.pendingAcks[peerID]...)
		cs.pendingAcks[peerID] = make([]uint32, 0)
	}
	if len(acks) == 0 {
		return nil
	}
	return &msgBatch{
		sessionID:  cs.sessionID,
		stateIndex: cs.stateIndex,
		id:         0, // Do not require an acknowledgement.
		src:        uint16(cs.impl.ID),
		dst:        peerID,
		msgs:       []*hbbft.ACSMessage{},
		acks:       acks,
	}
}

func (cs *CommonSubset) send(msgBatch *msgBatch) {
	if msgBatch == nil {
		// makeAckOnlyBatch can produce nil batches, if there is nothing
		// to acknowledge. We handle that here, to avoid IFs in multiple places.
		return
	}
	cs.log.Debugf("ACS::IO - Sending a msgBatch=%+v", msgBatch)
	cs.committeePeerGroup.SendMsgByIndex(msgBatch.dst, peering.PeerMessageReceiverCommonSubset, peerMsgTypeBatch, msgBatch.Bytes())
}

// endregion ///////////////////////////////////////////////////////////////////

// region msgBatch /////////////////////////////////////////////////////////////

/** msgBatch groups messages generated at one step in the protocol for a single recipient. */
type msgBatch struct {
	sessionID  uint64
	stateIndex uint32
	id         uint32              // ID of the batch for the acks, ID=0 => Acks are not needed.
	src        uint16              // Sender of the batch.
	dst        uint16              // Recipient of the batch.
	msgs       []*hbbft.ACSMessage // New messages to send.
	acks       []uint32            // Acknowledgements.
}

const (
	acsMsgTypeRbcProofRequest byte = 1 << 4
	acsMsgTypeRbcEchoRequest  byte = 2 << 4
	acsMsgTypeRbcReadyRequest byte = 3 << 4
	acsMsgTypeAbaBvalRequest  byte = 4 << 4
	acsMsgTypeAbaAuxRequest   byte = 5 << 4
	acsMsgTypeAbaCCRequest    byte = 6 << 4
	acsMsgTypeAbaDoneRequest  byte = 7 << 4
)

func newMsgBatch(data []byte) (*msgBatch, error) {
	mb := &msgBatch{}
	r := bytes.NewReader(data)
	if err := mb.Read(r); err != nil {
		return nil, err
	}
	return mb, nil
}

func (b *msgBatch) NeedsAck() bool {
	return b.id != 0
}

//nolint:funlen, gocyclo // TODO this function is too long and has a high cyclomatic complexity, should be refactored
func (b *msgBatch) Write(w io.Writer) error {
	var err error
	if err = util.WriteUint64(w, b.sessionID); err != nil {
		return xerrors.Errorf("failed to write msgBatch.sessionID: %w", err)
	}
	if err = util.WriteUint32(w, b.stateIndex); err != nil {
		return xerrors.Errorf("failed to write msgBatch.stateIndex: %w", err)
	}
	if err = util.WriteUint32(w, b.id); err != nil {
		return xerrors.Errorf("failed to write msgBatch.id: %w", err)
	}
	if err = util.WriteUint16(w, b.src); err != nil {
		return xerrors.Errorf("failed to write msgBatch.src: %w", err)
	}
	if err = util.WriteUint16(w, b.dst); err != nil {
		return xerrors.Errorf("failed to write msgBatch.dst: %w", err)
	}
	if err = util.WriteUint16(w, uint16(len(b.msgs))); err != nil {
		return xerrors.Errorf("failed to write msgBatch.msgs.len: %w", err)
	}
	for i, acsMsg := range b.msgs {
		if err = util.WriteUint16(w, uint16(acsMsg.ProposerID)); err != nil {
			return xerrors.Errorf("failed to write msgBatch.msgs[%v].ProposerID: %w", i, err)
		}
		switch acsMsgPayload := acsMsg.Payload.(type) {
		case *hbbft.BroadcastMessage:
			switch rbcMsgPayload := acsMsgPayload.Payload.(type) {
			case *hbbft.ProofRequest:
				if err = util.WriteByte(w, acsMsgTypeRbcProofRequest); err != nil {
					return xerrors.Errorf("failed to write msgBatch.msgs[%v].type: %w", i, err)
				}
				if err = util.WriteBytes16(w, rbcMsgPayload.RootHash); err != nil {
					return xerrors.Errorf("failed to write msgBatch.msgs[%v].RootHash: %w", i, err)
				}
				if err = util.WriteUint32(w, uint32(len(rbcMsgPayload.Proof))); err != nil {
					return xerrors.Errorf("failed to write msgBatch.msgs[%v].Proof.len: %w", i, err)
				}
				for pi, p := range rbcMsgPayload.Proof {
					if err = util.WriteBytes32(w, p); err != nil {
						return xerrors.Errorf("failed to write msgBatch.msgs[%v].Proof[%v]: %w", i, pi, err)
					}
				}
				if err = util.WriteUint16(w, uint16(rbcMsgPayload.Index)); err != nil {
					return xerrors.Errorf("failed to write msgBatch.msgs[%v].Index: %w", i, err)
				}
				if err = util.WriteUint16(w, uint16(rbcMsgPayload.Leaves)); err != nil {
					return xerrors.Errorf("failed to write msgBatch.msgs[%v].Leaves: %w", i, err)
				}
			case *hbbft.EchoRequest:
				if err = util.WriteByte(w, acsMsgTypeRbcEchoRequest); err != nil {
					return xerrors.Errorf("failed to write msgBatch.msgs[%v].type: %w", i, err)
				}
				if err = util.WriteBytes16(w, rbcMsgPayload.RootHash); err != nil {
					return xerrors.Errorf("failed to write msgBatch.msgs[%v].RootHash: %w", i, err)
				}
				if err = util.WriteUint32(w, uint32(len(rbcMsgPayload.Proof))); err != nil {
					return xerrors.Errorf("failed to write msgBatch.msgs[%v].Proof.len: %w", i, err)
				}
				for pi, p := range rbcMsgPayload.Proof {
					if err = util.WriteBytes32(w, p); err != nil {
						return xerrors.Errorf("failed to write msgBatch.msgs[%v].Proof[%v]: %w", i, pi, err)
					}
				}
				if err = util.WriteUint16(w, uint16(rbcMsgPayload.Index)); err != nil {
					return xerrors.Errorf("failed to write msgBatch.msgs[%v].Index: %w", i, err)
				}
				if err = util.WriteUint16(w, uint16(rbcMsgPayload.Leaves)); err != nil {
					return xerrors.Errorf("failed to write msgBatch.msgs[%v].Leaves: %w", i, err)
				}
			case *hbbft.ReadyRequest:
				if err = util.WriteByte(w, acsMsgTypeRbcReadyRequest); err != nil {
					return xerrors.Errorf("failed to write msgBatch.msgs[%v].type: %w", i, err)
				}
				if err = util.WriteBytes16(w, rbcMsgPayload.RootHash); err != nil {
					return xerrors.Errorf("failed to write msgBatch.msgs[%v].RootHash: %w", i, err)
				}
			default:
				return xerrors.Errorf("failed to write msgBatch.msgs[%v]: unexpected broadcast message type", i)
			}
		case *hbbft.AgreementMessage:
			switch abaMsgPayload := acsMsgPayload.Message.(type) {
			case *hbbft.BvalRequest:
				encoded := acsMsgTypeAbaBvalRequest
				if abaMsgPayload.Value {
					encoded |= 1
				}
				if err = util.WriteByte(w, encoded); err != nil {
					return xerrors.Errorf("failed to write msgBatch.msgs[%v].value: %w", i, err)
				}
				if err = util.WriteUint16(w, uint16(acsMsgPayload.Epoch)); err != nil {
					return xerrors.Errorf("failed to write msgBatch.msgs[%v].epoch: %w", i, err)
				}
			case *hbbft.AuxRequest:
				encoded := acsMsgTypeAbaAuxRequest
				if abaMsgPayload.Value {
					encoded |= 1
				}
				if err = util.WriteByte(w, encoded); err != nil {
					return xerrors.Errorf("failed to write msgBatch.msgs[%v].value: %w", i, err)
				}
				if err = util.WriteUint16(w, uint16(acsMsgPayload.Epoch)); err != nil {
					return xerrors.Errorf("failed to write msgBatch.msgs[%v].epoch: %w", i, err)
				}
			case *hbbft.CCRequest:
				coinMsg := abaMsgPayload.Payload.(*commoncoin.BlsCommonCoinMsg)
				if err = util.WriteByte(w, acsMsgTypeAbaCCRequest); err != nil {
					return xerrors.Errorf("failed to write msgBatch.msgs[%v].type: %w", i, err)
				}
				if err = util.WriteUint16(w, uint16(acsMsgPayload.Epoch)); err != nil {
					return xerrors.Errorf("failed to write msgBatch.msgs[%v].epoch: %w", i, err)
				}
				if err = coinMsg.Write(w); err != nil {
					return xerrors.Errorf("failed to write msgBatch.msgs[%v].Payload: %w", i, err)
				}
			case *hbbft.DoneRequest:
				encoded := acsMsgTypeAbaDoneRequest
				if err = util.WriteByte(w, encoded); err != nil {
					return xerrors.Errorf("failed to write msgBatch.msgs[%v].type: %w", i, err)
				}
				if err = util.WriteUint16(w, uint16(acsMsgPayload.Epoch)); err != nil {
					return xerrors.Errorf("failed to write msgBatch.msgs[%v].epoch: %w", i, err)
				}
			default:
				return xerrors.Errorf("failed to write msgBatch.msgs[%v]: unexpected agreemet message type", i)
			}
		default:
			return xerrors.Errorf("failed to write msgBatch.msgs[%v]: unexpected acs message type", i)
		}
	}
	if err = util.WriteUint16(w, uint16(len(b.acks))); err != nil {
		return xerrors.Errorf("failed to write msgBatch.acks.len: %w", err)
	}
	for i, ack := range b.acks {
		if err = util.WriteUint32(w, ack); err != nil {
			return xerrors.Errorf("failed to write msgBatch.acks[%v]: %w", i, err)
		}
	}
	return nil
}

//nolint:funlen, gocyclo // TODO this function is too long and has a high cyclomatic complexity, should be refactored
func (b *msgBatch) Read(r io.Reader) error {
	var err error
	if err = util.ReadUint64(r, &b.sessionID); err != nil {
		return xerrors.Errorf("failed to read msgBatch.sessionID: %w", err)
	}
	if err = util.ReadUint32(r, &b.stateIndex); err != nil {
		return xerrors.Errorf("failed to read msgBatch.stateIndex: %w", err)
	}
	if err = util.ReadUint32(r, &b.id); err != nil {
		return xerrors.Errorf("failed to read msgBatch.id: %w", err)
	}
	if err = util.ReadUint16(r, &b.src); err != nil {
		return xerrors.Errorf("failed to read msgBatch.src: %w", err)
	}
	if err = util.ReadUint16(r, &b.dst); err != nil {
		return xerrors.Errorf("failed to read msgBatch.dst: %w", err)
	}
	//
	// Msgs.
	var msgsLen uint16
	if err = util.ReadUint16(r, &msgsLen); err != nil {
		return xerrors.Errorf("failed to read msgBatch.msgs.len: %w", err)
	}
	b.msgs = make([]*hbbft.ACSMessage, msgsLen)
	for mi := range b.msgs {
		acsMsg := hbbft.ACSMessage{}
		b.msgs[mi] = &acsMsg
		// ProposerID.
		var proposerID uint16
		if err = util.ReadUint16(r, &proposerID); err != nil {
			return xerrors.Errorf("failed to read msgBatch.msgs[%v].ProposerID: %w", mi, err)
		}
		acsMsg.ProposerID = uint64(proposerID)
		// By Type.
		var msgType byte
		if msgType, err = util.ReadByte(r); err != nil {
			return xerrors.Errorf("failed to read msgBatch.msgs[%v].type: %w", mi, err)
		}
		switch msgType & 0xF0 {
		case acsMsgTypeRbcProofRequest:
			proofRequest := hbbft.ProofRequest{}
			if proofRequest.RootHash, err = util.ReadBytes16(r); err != nil {
				return xerrors.Errorf("failed to read msgBatch.msgs[%v].RootHash: %w", mi, err)
			}
			var proofLen uint32
			if err = util.ReadUint32(r, &proofLen); err != nil {
				return xerrors.Errorf("failed to read msgBatch.msgs[%v].Proof.len: %w", mi, err)
			}
			proofRequest.Proof = make([][]byte, proofLen)
			for pi := range proofRequest.Proof {
				if proofRequest.Proof[pi], err = util.ReadBytes32(r); err != nil {
					return xerrors.Errorf("failed to read msgBatch.msgs[%v].Proof[%v]: %w", mi, pi, err)
				}
			}
			var proofRequestIndex uint16
			if err = util.ReadUint16(r, &proofRequestIndex); err != nil {
				return xerrors.Errorf("failed to read msgBatch.msgs[%v].Index: %w", mi, err)
			}
			proofRequest.Index = int(proofRequestIndex)
			var proofRequestLeaves uint16
			if err = util.ReadUint16(r, &proofRequestLeaves); err != nil {
				return xerrors.Errorf("failed to read msgBatch.msgs[%v].Leaves: %w", mi, err)
			}
			proofRequest.Leaves = int(proofRequestLeaves)
			acsMsg.Payload = &hbbft.BroadcastMessage{
				Payload: &proofRequest,
			}
		case acsMsgTypeRbcEchoRequest:
			echoRequest := hbbft.EchoRequest{}
			if echoRequest.RootHash, err = util.ReadBytes16(r); err != nil {
				return xerrors.Errorf("failed to read msgBatch.msgs[%v].RootHash: %w", mi, err)
			}
			var proofLen uint32
			if err = util.ReadUint32(r, &proofLen); err != nil {
				return xerrors.Errorf("failed to read msgBatch.msgs[%v].Proof.len: %w", mi, err)
			}
			echoRequest.Proof = make([][]byte, proofLen)
			for pi := range echoRequest.Proof {
				if echoRequest.Proof[pi], err = util.ReadBytes32(r); err != nil {
					return xerrors.Errorf("failed to read msgBatch.msgs[%v].Proof[%v]: %w", mi, pi, err)
				}
			}
			var echoRequestIndex uint16
			if err = util.ReadUint16(r, &echoRequestIndex); err != nil {
				return xerrors.Errorf("failed to read msgBatch.msgs[%v].Index: %w", mi, err)
			}
			echoRequest.Index = int(echoRequestIndex)
			var echoRequestLeaves uint16
			if err = util.ReadUint16(r, &echoRequestLeaves); err != nil {
				return xerrors.Errorf("failed to read msgBatch.msgs[%v].Leaves: %w", mi, err)
			}
			echoRequest.Leaves = int(echoRequestLeaves)
			acsMsg.Payload = &hbbft.BroadcastMessage{
				Payload: &echoRequest,
			}
		case acsMsgTypeRbcReadyRequest:
			readyRequest := hbbft.ReadyRequest{}
			if readyRequest.RootHash, err = util.ReadBytes16(r); err != nil {
				return xerrors.Errorf("failed to read msgBatch.msgs[%v].RootHash: %w", mi, err)
			}
			acsMsg.Payload = &hbbft.BroadcastMessage{
				Payload: &readyRequest,
			}
		case acsMsgTypeAbaBvalRequest:
			var epoch uint16
			if err = util.ReadUint16(r, &epoch); err != nil {
				return xerrors.Errorf("failed to read msgBatch.msgs[%v].epoch: %w", mi, err)
			}
			acsMsg.Payload = &hbbft.AgreementMessage{
				Epoch:   int(epoch),
				Message: &hbbft.BvalRequest{Value: msgType&0x01 == 1},
			}
		case acsMsgTypeAbaAuxRequest:
			var epoch uint16
			if err = util.ReadUint16(r, &epoch); err != nil {
				return xerrors.Errorf("failed to read msgBatch.msgs[%v].epoch: %w", mi, err)
			}
			acsMsg.Payload = &hbbft.AgreementMessage{
				Epoch:   int(epoch),
				Message: &hbbft.AuxRequest{Value: msgType&0x01 == 1},
			}
		case acsMsgTypeAbaCCRequest:
			var epoch uint16
			if err = util.ReadUint16(r, &epoch); err != nil {
				return xerrors.Errorf("failed to read msgBatch.msgs[%v].epoch: %w", mi, err)
			}
			var ccReq commoncoin.BlsCommonCoinMsg
			if err = ccReq.Read(r); err != nil {
				return xerrors.Errorf("failed to read msgBatch.msgs[%v].Payload: %w", mi, err)
			}
			acsMsg.Payload = &hbbft.AgreementMessage{
				Epoch: int(epoch),
				Message: &hbbft.CCRequest{
					Payload: &ccReq,
				},
			}
		case acsMsgTypeAbaDoneRequest:
			var epoch uint16
			if err = util.ReadUint16(r, &epoch); err != nil {
				return xerrors.Errorf("failed to read msgBatch.msgs[%v].epoch: %w", mi, err)
			}
			acsMsg.Payload = &hbbft.AgreementMessage{
				Epoch:   int(epoch),
				Message: &hbbft.DoneRequest{},
			}
		default:
			return xerrors.Errorf("failed to read msgBatch.msgs[%v]: unexpected message type %v, b=%+v", mi, msgType, b)
		}
	}
	//
	// Acks.
	var acksLen uint16
	if err = util.ReadUint16(r, &acksLen); err != nil {
		return xerrors.Errorf("failed to read msgBatch.acks.len: %w", err)
	}
	b.acks = make([]uint32, acksLen)
	for ai := range b.acks {
		if err = util.ReadUint32(r, &b.acks[ai]); err != nil {
			return xerrors.Errorf("failed to read msgBatch.acks[%v]: %w", ai, err)
		}
	}
	return nil
}

func (b *msgBatch) Bytes() []byte {
	var buf bytes.Buffer
	_ = b.Write(&buf)
	return buf.Bytes()
}

// endregion ///////////////////////////////////////////////////////////////////
