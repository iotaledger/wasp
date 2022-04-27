// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

import (
	"time"

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

func (a *ackHandler) Input(input Input) []Message {
	return a.makeBatches(a.nested.Input(input))
}

func (a *ackHandler) Message(msg Message) []Message {
	switch msgT := msg.(type) {
	case *ackHandlerTick:
		return a.handleTickMsg(msgT)
	case *ackHandlerBatch:
		return a.handleBatchMsg(msgT)
	default:
		panic(xerrors.Errorf("unexpected message type: %+v", msg))
	}
}

func (a *ackHandler) Output() Output {
	return a.nested.Output()
}

func (a *ackHandler) handleTickMsg(msg *ackHandlerTick) []Message {
	resendOlderThan := msg.timestamp.Add(-a.resendPeriod)
	resendMsgs := []Message{}
	for _, nodeSentUnacked := range a.sentUnacked {
		for batchID, batch := range nodeSentUnacked {
			if batch.sent == nil {
				// Don't resent, just mark the current timestamp.
				// We have sent it after the previous tick.
				batch.sent = &msg.timestamp
			} else if batch.sent.Before(resendOlderThan) {
				// Resent it, timeout is already passed.
				batch.sent = &msg.timestamp
				resendMsgs = append(resendMsgs, nodeSentUnacked[batchID])
			}
		}
	}
	return resendMsgs
}

func (a *ackHandler) handleBatchMsg(msgBatch *ackHandlerBatch) []Message {
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
			return []Message{
				&ackHandlerBatch{
					recipient: msgBatch.sender,
					id:        nil,                 // That's ack-only.
					msgs:      []Message{},         // No payload.
					acks:      []int{*msgBatch.id}, // Ack single message.
					sent:      nil,                 // We will not track this message, it has no payload.
				},
			}
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
		return []Message{ackedBatch}
	}
	//
	// That's new batch, we have to process it.
	nestedMsgs := []Message{}
	for i := range msgBatch.msgs {
		nestedMsgs = append(nestedMsgs, a.nested.Message(msgBatch.msgs[i])...)
	}
	if _, ok := a.recvAcksIn[msgBatch.sender]; !ok {
		a.recvAcksIn[msgBatch.sender] = map[int]*int{}
	}
	a.recvAcksIn[msgBatch.sender][*msgBatch.id] = nil
	return a.makeBatches(nestedMsgs)
}

func (a *ackHandler) makeBatches(msgs []Message) []Message {
	groupedMsgs := map[NodeID][]Message{}
	for i := range msgs {
		msgRecipient := msgs[i].Recipient()
		if recipientMsgs, ok := groupedMsgs[msgRecipient]; ok {
			groupedMsgs[msgRecipient] = append(recipientMsgs, msgs[i])
		} else {
			groupedMsgs[msgRecipient] = []Message{msgs[i]}
		}
	}
	batches := []Message{}
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

		batches = append(batches, batch)
	}
	return batches
}

//
//	Message conveying the message batches and acknowledgements.
//
type ackHandlerBatch struct {
	sender    NodeID
	recipient NodeID
	id        *int       // That's ACK only, if nil.
	msgs      []Message  // Messages in the batch.
	acks      []int      // Acknowledged batches.
	sent      *time.Time // Transient, only used for outgoing messages, not sent to the outside.
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
	panic("implement!") // TODO: ...
}

//
// Event representing a timer tick.
//
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
