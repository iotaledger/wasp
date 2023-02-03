// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

import (
	"bytes"
	"fmt"
	"time"

	"github.com/iotaledger/wasp/packages/util"
)

const (
	ackHandlerMsgTypeReset byte = iota
	ackHandlerMsgTypeBatch
)

// The purpose of this wrapper is to handle unreliable network by implementing
// a RELIABLE CHANNEL abstraction. This is done by resending messages until an
// acknowledgement is received. To make this more efficient, acknowledgements
// are piggy-backed on other messages (or sent stand-alone, if there is no
// messages to piggy-back the acknowledgements).
type ackHandler struct {
	me           NodeID
	nested       GPA
	initialized  map[NodeID]bool
	initPending  map[NodeID][]Message
	counters     map[NodeID]int // For numbering the outgoing messages.
	resendPeriod time.Duration
	sentUnacked  map[NodeID]map[int]*ackHandlerBatch
	recvAcksIn   map[NodeID]map[int]*int
}

type AckHandler interface {
	GPA
	DismissPeer(peerID NodeID) // To avoid resending messages to dead peers.
	MakeTickInput(time.Time) Input
	NestedMessage(msg Message) OutMessages
	NestedCall(c func(GPA) OutMessages) OutMessages
}

var _ AckHandler = &ackHandler{}

func NewAckHandler(me NodeID, nested GPA, resendPeriod time.Duration) AckHandler {
	return &ackHandler{
		me:           me,
		nested:       nested,
		initialized:  map[NodeID]bool{},
		initPending:  map[NodeID][]Message{},
		counters:     map[NodeID]int{},
		resendPeriod: resendPeriod,
		sentUnacked:  map[NodeID]map[int]*ackHandlerBatch{},
		recvAcksIn:   map[NodeID]map[int]*int{},
	}
}

func (a *ackHandler) DismissPeer(peerID NodeID) {
	delete(a.initialized, peerID)
	delete(a.initPending, peerID)
	delete(a.counters, peerID)
	delete(a.sentUnacked, peerID)
	delete(a.recvAcksIn, peerID)
}

func (a *ackHandler) MakeTickInput(timestamp time.Time) Input {
	return &ackHandlerTick{timestamp: timestamp}
}

func (a *ackHandler) Input(input Input) OutMessages {
	switch input := input.(type) {
	case *ackHandlerTick:
		return a.handleTickMsg(input)
	default:
		return a.makeBatches(a.nested.Input(input))
	}
}

func (a *ackHandler) Message(msg Message) OutMessages {
	switch msg := msg.(type) {
	case *ackHandlerReset:
		return a.handleResetMsg(msg)
	case *ackHandlerBatch:
		return a.handleBatchMsg(msg)
	default:
		panic(fmt.Errorf("unexpected message type: %+v", msg))
	}
}

func (a *ackHandler) NestedMessage(msg Message) OutMessages {
	return a.makeBatches(a.nested.Message(msg))
}

func (a *ackHandler) NestedCall(c func(GPA) OutMessages) OutMessages {
	return a.makeBatches(c(a.nested))
}

func (a *ackHandler) Output() Output {
	return a.nested.Output()
}

func (a *ackHandler) StatusString() string {
	return fmt.Sprintf("{ACK:%s}", a.nested.StatusString())
}

func (a *ackHandler) UnmarshalMessage(data []byte) (Message, error) {
	if len(data) < 1 {
		return nil, fmt.Errorf("Data to short: %v", data)
	}
	switch data[0] {
	case ackHandlerMsgTypeReset:
		msg := &ackHandlerReset{}
		if err := msg.UnmarshalBinary(data); err != nil {
			return nil, fmt.Errorf("cannot unmarshal ackHandlerReset: %w", err)
		}
		return msg, nil
	case ackHandlerMsgTypeBatch:
		msg := &ackHandlerBatch{nestedGPA: a.nested}
		if err := msg.UnmarshalBinary(data); err != nil {
			return nil, fmt.Errorf("cannot unmarshal ackHandlerBatch: %w", err)
		}
		return msg, nil
	default:
		return nil, fmt.Errorf("unexpected message type: %v", data[0])
	}
}

func (a *ackHandler) handleTickMsg(msg *ackHandlerTick) OutMessages {
	resendOlderThan := msg.timestamp.Add(-a.resendPeriod)
	resendMsgs := NoMessages()
	for _, nodeSentUnacked := range a.sentUnacked {
		for batchID, batch := range nodeSentUnacked {
			if batch.sent == nil {
				// Don't resent, just mark the current timestamp.
				// We have sent it after the previous tick.
				batch.sent = &msg.timestamp
			} else if batch.sent.Before(resendOlderThan) {
				// Resent it, timeout is already passed.
				batch.sent = &msg.timestamp
				resendMsgs.Add(nodeSentUnacked[batchID])
			}
		}
	}
	for nodeID := range a.initPending {
		resendMsgs.Add(&ackHandlerReset{BasicMessage: NewBasicMessage(nodeID), response: false, latestID: 0})
	}
	return resendMsgs
}

func (a *ackHandler) handleResetMsg(msg *ackHandlerReset) OutMessages {
	from := msg.sender
	if !msg.response {
		max := 0
		if recvAcksIn, ok := a.recvAcksIn[msg.sender]; ok {
			for id := range recvAcksIn {
				if id > max {
					max = id
				}
			}
		}
		return NoMessages().Add(&ackHandlerReset{
			BasicMessage: NewBasicMessage(msg.sender),
			response:     true,
			latestID:     max,
		})
	}
	if ini, ok := a.initialized[from]; ok && ini {
		return nil
	}
	a.counters[msg.sender] = msg.latestID + 1
	a.initialized[msg.sender] = true
	return a.makeBatches(NoMessages())
}

func (a *ackHandler) handleBatchMsg(msgBatch *ackHandlerBatch) OutMessages {
	//
	// Process the received acknowledgements.
	// Drop all the outgoing batches, that are now acknowledged.
	for _, ackedBatchID := range msgBatch.acks {
		if unacked, ok := a.sentUnacked[msgBatch.sender]; ok {
			delete(unacked, ackedBatchID)
		}
	}
	//
	// Was that ack-only message?
	if msgBatch.id == nil {
		// That was ack-only batch, nothing more to do with it.
		return NoMessages()
	}

	peerRecvAcksIn, ok := a.recvAcksIn[msgBatch.sender]
	if !ok {
		peerRecvAcksIn = map[int]*int{}
	}
	batchAckedIn, ok := peerRecvAcksIn[*msgBatch.id]
	if ok {
		// Was received already before.
		if batchAckedIn == nil {
			// Not acknowledged yet, just send an ack-only message for now.
			// The sender has already re-sent the message, so it waits for the ack.
			return NoMessages().Add(&ackHandlerBatch{
				recipient: msgBatch.sender,
				id:        nil,                 // That's ack-only.
				msgs:      []Message{},         // No payload.
				acks:      []int{*msgBatch.id}, // Ack single message.
				sent:      nil,                 // We will not track this message, it has no payload.
			})
		}
		//
		// We have acked it already. If we have the batch with an ack, we
		// resent it. Otherwise the ack was already acked and this message
		// is outdated and can be ignored.
		peerSentUnacked, ok := a.sentUnacked[msgBatch.sender]
		if !ok {
			return NoMessages()
		}
		ackedBatch, ok := peerSentUnacked[*batchAckedIn]
		if !ok {
			return NoMessages()
		}
		now := time.Now()
		ackedBatch.sent = &now
		return NoMessages().Add(ackedBatch)
	}
	//
	// That's new batch, we have to process it.
	nestedMsgs := NoMessages()
	for i := range msgBatch.msgs {
		nestedMsgs.AddAll(a.nested.Message(msgBatch.msgs[i]))
	}
	if _, ok := a.recvAcksIn[msgBatch.sender]; !ok {
		a.recvAcksIn[msgBatch.sender] = map[int]*int{}
	}
	a.recvAcksIn[msgBatch.sender][*msgBatch.id] = nil
	return a.makeBatches(nestedMsgs)
}

func (a *ackHandler) makeBatches(msgs OutMessages) OutMessages {
	if msgs == nil {
		return nil
	}
	groupedMsgs := map[NodeID][]Message{}
	msgs.MustIterate(func(msg Message) {
		msgRecipient := msg.Recipient()
		if recipientMsgs, ok := groupedMsgs[msgRecipient]; ok {
			groupedMsgs[msgRecipient] = append(recipientMsgs, msg)
		} else {
			groupedMsgs[msgRecipient] = []Message{msg}
		}
	})
	for nodeID, pending := range a.initPending {
		if gr, ok := groupedMsgs[nodeID]; ok {
			groupedMsgs[nodeID] = append(gr, pending...)
		} else {
			groupedMsgs[nodeID] = pending
		}
	}
	a.initPending = map[NodeID][]Message{}

	batches := NoMessages()
	for nodeID, batchMsgs := range groupedMsgs {
		if initialized, ok := a.initialized[nodeID]; !ok || !initialized {
			if pending, ok := a.initPending[nodeID]; ok {
				a.initPending[nodeID] = append(pending, batchMsgs...)
			} else {
				a.initPending[nodeID] = batchMsgs
			}
			batches.Add(&ackHandlerReset{BasicMessage: NewBasicMessage(nodeID), response: false, latestID: 0})
			continue
		}
		//
		// Assign batch ID.
		if _, ok := a.counters[nodeID]; !ok {
			a.counters[nodeID] = 0
		}
		batchID := a.counters[nodeID]
		a.counters[nodeID]++
		//
		// Collect batches to be acknowledged and mark them as acknowledged.
		acks := []int{}
		if nodeRecvAcksIn, ok := a.recvAcksIn[nodeID]; ok {
			for recvBatchID, ackedIn := range nodeRecvAcksIn {
				if ackedIn == nil {
					acks = append(acks, recvBatchID)
					nodeRecvAcksIn[recvBatchID] = &batchID
				}
			}
		}
		//
		// Produce the batch and register it as unacked.
		batch := &ackHandlerBatch{
			sender:    a.me,
			recipient: nodeID,
			id:        &batchID,
			acks:      acks,
			msgs:      batchMsgs,
			sent:      nil, // Will be set after first resend, to avoid resend to early.
		}
		if _, ok := a.sentUnacked[nodeID]; !ok {
			a.sentUnacked[nodeID] = map[int]*ackHandlerBatch{*batch.id: batch}
		} else {
			a.sentUnacked[nodeID][*batch.id] = batch
		}

		batches.Add(batch)
	}
	return batches
}

////////////////////////////////////////////////////////////////////////////////
// ackHandlerReset

type ackHandlerReset struct {
	BasicMessage
	response bool
	latestID int
}

var _ Message = &ackHandlerReset{}

func (m *ackHandlerReset) MarshalBinary() ([]byte, error) {
	w := &bytes.Buffer{}
	if err := util.WriteByte(w, ackHandlerMsgTypeReset); err != nil {
		return nil, fmt.Errorf("cannot serialize ackHandlerReset.msgType: %w", err)
	}
	if err := util.WriteBoolByte(w, m.response); err != nil {
		return nil, fmt.Errorf("cannot serialize ackHandlerReset.response: %w", err)
	}
	if err := util.WriteUint32(w, uint32(m.latestID)); err != nil {
		return nil, fmt.Errorf("cannot serialize ackHandlerReset.latestID: %w", err)
	}
	return w.Bytes(), nil
}

func (m *ackHandlerReset) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)
	msgType, err := util.ReadByte(r)
	if err != nil {
		return fmt.Errorf("cannot deserialize ackHandlerReset.msgType: %w", err)
	}
	if msgType != ackHandlerMsgTypeReset {
		return fmt.Errorf("unexpected msgType: %v", msgType)
	}
	if err := util.ReadBoolByte(r, &m.response); err != nil {
		return fmt.Errorf("cannot deserialize ackHandlerReset.response: %w", err)
	}
	var u32 uint32
	if err := util.ReadUint32(r, &u32); err != nil {
		return fmt.Errorf("cannot deserialize ackHandlerReset.latestID: %w", err)
	}
	m.latestID = int(u32)
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// ackHandlerBatch

// Message conveying the message batches and acknowledgements.
type ackHandlerBatch struct {
	sender    NodeID
	recipient NodeID
	id        *int       // That's ACK only, if nil.
	msgs      []Message  // Messages in the batch.
	acks      []int      // Acknowledged batches.
	sent      *time.Time // Transient, only used for outgoing messages, not sent to the outside.
	nestedGPA GPA        // Transient, for un-marshaling only.
}

var _ Message = &ackHandlerBatch{}

func (m *ackHandlerBatch) Recipient() NodeID {
	return m.recipient
}

func (m *ackHandlerBatch) SetSender(sender NodeID) {
	m.sender = sender
	for _, msg := range m.msgs {
		msg.SetSender(sender)
	}
}

func (m *ackHandlerBatch) MarshalBinary() ([]byte, error) {
	w := &bytes.Buffer{}
	if err := util.WriteByte(w, ackHandlerMsgTypeBatch); err != nil {
		return nil, fmt.Errorf("cannot serialize ackHandlerBatch.msgType: %w", err)
	}
	//
	// m.id
	if m.id != nil {
		if err := util.WriteBoolByte(w, true); err != nil {
			return nil, fmt.Errorf("cannot serialize ackHandlerBatch.id!=nil: %w", err)
		}
		if err := util.WriteUint32(w, uint32(*m.id)); err != nil {
			return nil, fmt.Errorf("cannot serialize ackHandlerBatch.id: %w", err)
		}
	} else {
		if err := util.WriteBoolByte(w, false); err != nil {
			return nil, fmt.Errorf("cannot serialize ackHandlerBatch.id==nil: %w", err)
		}
	}
	//
	// m.msgs
	if err := util.WriteUint16(w, uint16(len(m.msgs))); err != nil {
		return nil, fmt.Errorf("cannot serialize ackHandlerBatch.msgs.length: %w", err)
	}
	for i := range m.msgs {
		msgData, err := m.msgs[i].MarshalBinary()
		if err != nil {
			return nil, fmt.Errorf("cannot serialize ackHandlerBatch.msgs[%v]: %w", i, err)
		}
		if err := util.WriteBytes32(w, msgData); err != nil {
			return nil, fmt.Errorf("cannot serialize ackHandlerBatch.msgs[%v]: %w", i, err)
		}
	}
	//
	// m.acks
	if err := util.WriteUint16(w, uint16(len(m.acks))); err != nil {
		return nil, fmt.Errorf("cannot serialize ackHandlerBatch.acks.length: %w", err)
	}
	for i := range m.acks {
		if err := util.WriteUint32(w, uint32(m.acks[i])); err != nil {
			return nil, fmt.Errorf("cannot serialize ackHandlerBatch.acks[%v]: %w", i, err)
		}
	}
	return w.Bytes(), nil
}

func (m *ackHandlerBatch) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)
	msgType, err := util.ReadByte(r)
	if err != nil {
		return fmt.Errorf("cannot deserialize ackHandlerBatch.msgType: %w", err)
	}
	if msgType != ackHandlerMsgTypeBatch {
		return fmt.Errorf("unexpected msgType: %v", msgType)
	}
	//
	// m.id
	var mIDPresent bool
	if err := util.ReadBoolByte(r, &mIDPresent); err != nil {
		return fmt.Errorf("cannot deserialize ackHandlerBatch.id?=nil: %w", err)
	}
	if mIDPresent {
		var mID uint32
		if err := util.ReadUint32(r, &mID); err != nil {
			return fmt.Errorf("cannot deserialize ackHandlerBatch.id: %w", err)
		}
		mIDasInt := int(mID)
		m.id = &mIDasInt
	} else {
		m.id = nil
	}
	//
	// m.msgs
	var msgsLen uint16
	if err := util.ReadUint16(r, &msgsLen); err != nil {
		return fmt.Errorf("cannot deserialize ackHandlerBatch.msgs.length: %w", err)
	}
	m.msgs = make([]Message, msgsLen)
	for i := range m.msgs {
		msgData, err := util.ReadBytes32(r)
		if err != nil {
			return fmt.Errorf("cannot deserialize ackHandlerBatch.msgs[%v]: %w", i, err)
		}
		msg, err := m.nestedGPA.UnmarshalMessage(msgData)
		if err != nil {
			return fmt.Errorf("cannot deserialize ackHandlerBatch.msgs[%v]: %w", i, err)
		}
		m.msgs[i] = msg
	}
	//
	// m.acks
	var acksLen uint16
	if err := util.ReadUint16(r, &acksLen); err != nil {
		return fmt.Errorf("cannot deserialize ackHandlerBatch.acks.length: %w", err)
	}
	m.acks = make([]int, acksLen)
	for i := range m.acks {
		var ackedID uint32
		if err := util.ReadUint32(r, &ackedID); err != nil {
			return fmt.Errorf("cannot deserialize ackHandlerBatch.acks[%v]: %w", i, err)
		}
		m.acks[i] = int(ackedID)
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// ackHandlerTick

// Event representing a timer tick.
type ackHandlerTick struct {
	timestamp time.Time
}
