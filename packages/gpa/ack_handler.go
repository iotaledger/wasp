// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

import (
	"bytes"
	"fmt"
	"time"

	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
)

// The purpose of this wrapper is to handle unreliable network by implementing
// a RELIABLE CHANNEL abstraction. This is done by resending messages until an
// acknowledgement is received. To make this more efficient, acknowledgements
// are piggy-backed on other messages (or sent stand-alone, if there is no
// messages to piggy-back the acknowledgements).
type ackHandler struct {
	me           NodeID
	nested       GPA
	counters     map[NodeID]int // For numbering the outgoing messages.
	resendPeriod time.Duration
	sentUnacked  map[NodeID]map[int]*ackHandlerBatch
	recvAcksIn   map[NodeID]map[int]*int
}

type AckHandler interface {
	GPA
	MakeTickMsg(time.Time) Message
	NestedMessage(msg Message) OutMessages
	NestedCall(c func(GPA) OutMessages) OutMessages
}

var _ AckHandler = &ackHandler{}

func NewAckHandler(me NodeID, nested GPA, resendPeriod time.Duration) AckHandler {
	return &ackHandler{
		me:           me,
		nested:       nested,
		counters:     map[NodeID]int{},
		resendPeriod: resendPeriod,
		sentUnacked:  map[NodeID]map[int]*ackHandlerBatch{},
		recvAcksIn:   map[NodeID]map[int]*int{},
	}
}

func (a *ackHandler) MakeTickMsg(timestamp time.Time) Message {
	return &ackHandlerTick{recipient: a.me, timestamp: timestamp}
}

func (a *ackHandler) Input(input Input) OutMessages {
	return a.makeBatches(a.nested.Input(input))
}

func (a *ackHandler) Message(msg Message) OutMessages {
	switch msgT := msg.(type) {
	case *ackHandlerTick:
		return a.handleTickMsg(msgT)
	case *ackHandlerBatch:
		return a.handleBatchMsg(msgT)
	default:
		panic(xerrors.Errorf("unexpected message type: %+v", msg))
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
	return fmt.Sprintf("{AckHandler, nested=%s}", a.nested.StatusString())
}

func (a *ackHandler) UnmarshalMessage(data []byte) (Message, error) {
	msg := &ackHandlerBatch{nestedGPA: a.nested}
	if err := msg.UnmarshalBinary(data); err != nil {
		return nil, xerrors.Errorf("cannot unmarshal ackHandlerBatch: %w", err)
	}
	return msg, nil
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
	return resendMsgs
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
	groupedMsgs := map[NodeID][]Message{}
	msgs.MustIterate(func(msg Message) {
		msgRecipient := msg.Recipient()
		if recipientMsgs, ok := groupedMsgs[msgRecipient]; ok {
			groupedMsgs[msgRecipient] = append(recipientMsgs, msg)
		} else {
			groupedMsgs[msgRecipient] = []Message{msg}
		}
	})

	batches := NoMessages()
	for nodeID, batchMsgs := range groupedMsgs {
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
	//
	// m.id
	if m.id != nil {
		if err := util.WriteBoolByte(w, true); err != nil {
			return nil, xerrors.Errorf("cannot serialize ackHandlerBatch.id!=nil: %w", err)
		}
		if err := util.WriteUint32(w, uint32(*m.id)); err != nil {
			return nil, xerrors.Errorf("cannot serialize ackHandlerBatch.id: %w", err)
		}
	} else {
		if err := util.WriteBoolByte(w, false); err != nil {
			return nil, xerrors.Errorf("cannot serialize ackHandlerBatch.id==nil: %w", err)
		}
	}
	//
	// m.msgs
	if err := util.WriteUint16(w, uint16(len(m.msgs))); err != nil {
		return nil, xerrors.Errorf("cannot serialize ackHandlerBatch.msgs.length: %w", err)
	}
	for i := range m.msgs {
		msgData, err := m.msgs[i].MarshalBinary()
		if err != nil {
			return nil, xerrors.Errorf("cannot serialize ackHandlerBatch.msgs[%v]: %w", i, err)
		}
		if err := util.WriteBytes32(w, msgData); err != nil {
			return nil, xerrors.Errorf("cannot serialize ackHandlerBatch.msgs[%v]: %w", i, err)
		}
	}
	//
	// m.acks
	if err := util.WriteUint16(w, uint16(len(m.acks))); err != nil {
		return nil, xerrors.Errorf("cannot serialize ackHandlerBatch.acks.length: %w", err)
	}
	for i := range m.acks {
		if err := util.WriteUint32(w, uint32(m.acks[i])); err != nil {
			return nil, xerrors.Errorf("cannot serialize ackHandlerBatch.acks[%v]: %w", i, err)
		}
	}
	return w.Bytes(), nil
}

func (m *ackHandlerBatch) UnmarshalBinary(data []byte) error {
	r := bytes.NewReader(data)
	//
	// m.id
	var mIDPresent bool
	if err := util.ReadBoolByte(r, &mIDPresent); err != nil {
		return xerrors.Errorf("cannot deserialize ackHandlerBatch.id?=nil: %w", err)
	}
	if mIDPresent {
		var mID uint32
		if err := util.ReadUint32(r, &mID); err != nil {
			return xerrors.Errorf("cannot deserialize ackHandlerBatch.id: %w", err)
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
		return xerrors.Errorf("cannot deserialize ackHandlerBatch.msgs.length: %w", err)
	}
	m.msgs = make([]Message, msgsLen)
	for i := range m.msgs {
		msgData, err := util.ReadBytes32(r)
		if err != nil {
			return xerrors.Errorf("cannot deserialize ackHandlerBatch.msgs[%v]: %w", i, err)
		}
		msg, err := m.nestedGPA.UnmarshalMessage(msgData)
		if err != nil {
			return xerrors.Errorf("cannot deserialize ackHandlerBatch.msgs[%v]: %w", i, err)
		}
		m.msgs[i] = msg
	}
	//
	// m.acks
	var acksLen uint16
	if err := util.ReadUint16(r, &acksLen); err != nil {
		return xerrors.Errorf("cannot deserialize ackHandlerBatch.acks.length: %w", err)
	}
	m.acks = make([]int, acksLen)
	for i := range m.acks {
		var ackedID uint32
		if err := util.ReadUint32(r, &ackedID); err != nil {
			return xerrors.Errorf("cannot deserialize ackHandlerBatch.acks[%v]: %w", i, err)
		}
		m.acks[i] = int(ackedID)
	}
	return nil
}

// Event representing a timer tick.
type ackHandlerTick struct {
	recipient NodeID
	timestamp time.Time
}

var _ Message = &ackHandlerTick{}

func (m *ackHandlerTick) Recipient() NodeID {
	return m.recipient
}

func (m *ackHandlerTick) SetSender(sender NodeID) {
	// Don't care the sender, that's local event.
}

func (m *ackHandlerTick) MarshalBinary() ([]byte, error) {
	panic(xerrors.Errorf("local event shouldn't be marshaled."))
}
