// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

import (
	"fmt"
	"time"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/hive.go/ds/shrinkingmap"
)

const (
	msgTypeAckHandlerReset MessageType = iota
	msgTypeAckHandlerBatch
)

// The purpose of this wrapper is to handle unreliable network by implementing
// a RELIABLE CHANNEL abstraction. This is done by resending messages until an
// acknowledgement is received. To make this more efficient, acknowledgements
// are piggy-backed on other messages (or sent stand-alone, if there is no
// messages to piggy-back the acknowledgements).
type ackHandler struct {
	me           NodeID
	nested       GPA
	resendPeriod time.Duration
	initialized  *shrinkingmap.ShrinkingMap[NodeID, bool]
	initPending  *shrinkingmap.ShrinkingMap[NodeID, []Message]
	counters     *shrinkingmap.ShrinkingMap[NodeID, int] // For numbering the outgoing messages.
	sentUnacked  *shrinkingmap.ShrinkingMap[NodeID, *shrinkingmap.ShrinkingMap[int, *ackHandlerBatch]]
	recvAcksIn   *shrinkingmap.ShrinkingMap[NodeID, map[int]*int]
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
		resendPeriod: resendPeriod,
		initialized:  shrinkingmap.New[NodeID, bool](),
		initPending:  shrinkingmap.New[NodeID, []Message](),
		counters:     shrinkingmap.New[NodeID, int](),
		sentUnacked:  shrinkingmap.New[NodeID, *shrinkingmap.ShrinkingMap[int, *ackHandlerBatch]](),
		recvAcksIn:   shrinkingmap.New[NodeID, map[int]*int](),
	}
}

func (a *ackHandler) DismissPeer(peerID NodeID) {
	a.initialized.Delete(peerID)
	a.initPending.Delete(peerID)
	a.counters.Delete(peerID)
	a.sentUnacked.Delete(peerID)
	a.recvAcksIn.Delete(peerID)
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
	msg, err := UnmarshalMessage(data, Mapper{
		msgTypeAckHandlerReset: func() Message { return &ackHandlerReset{} },
		msgTypeAckHandlerBatch: func() Message { return &ackHandlerBatch{nestedGPA: a.nested} },
	})
	if err != nil {
		fmt.Printf("ack, err=%v\n", err) // TODO: Clean this up.
	}
	return msg, err
}

func (a *ackHandler) handleTickMsg(msg *ackHandlerTick) OutMessages {
	resendOlderThan := msg.timestamp.Add(-a.resendPeriod)
	resendMsgs := NoMessages()
	a.sentUnacked.ForEach(func(_ NodeID, nodeSentUnacked *shrinkingmap.ShrinkingMap[int, *ackHandlerBatch]) bool {
		nodeSentUnacked.ForEach(func(batchID int, batch *ackHandlerBatch) bool {
			if batch.sent == nil {
				// Don't resent, just mark the current timestamp.
				// We have sent it after the previous tick.
				batch.sent = &msg.timestamp
			} else if batch.sent.Before(resendOlderThan) {
				// Resent it, timeout is already passed.
				batch.sent = &msg.timestamp
				resendMsgs.Add(batch)
			}

			return true
		})
		return true
	})

	a.initPending.ForEachKey(func(nodeID NodeID) bool {
		resendMsgs.Add(&ackHandlerReset{BasicMessage: NewBasicMessage(nodeID), response: false, latestID: 0})
		return true
	})
	return resendMsgs
}

func (a *ackHandler) handleResetMsg(msg *ackHandlerReset) OutMessages {
	from := msg.sender
	if !msg.response {
		maxID := 0

		if recvAcksIn, exists := a.recvAcksIn.Get(msg.sender); exists {
			for id := range recvAcksIn {
				if id > maxID {
					maxID = id
				}
			}
		}
		return NoMessages().Add(&ackHandlerReset{
			BasicMessage: NewBasicMessage(msg.sender),
			response:     true,
			latestID:     maxID,
		})
	}
	if ini, exists := a.initialized.Get(from); exists && ini {
		return nil
	}
	a.counters.Set(msg.sender, msg.latestID+1)
	a.initialized.Set(msg.sender, true)
	return a.makeBatches(NoMessages())
}

func (a *ackHandler) handleBatchMsg(msgBatch *ackHandlerBatch) OutMessages {
	//
	// Process the received acknowledgements.
	// Drop all the outgoing batches, that are now acknowledged.
	for _, ackedBatchID := range msgBatch.acks {
		if unacked, exists := a.sentUnacked.Get(msgBatch.sender); exists {
			unacked.Delete(ackedBatchID)
		}
	}
	//
	// Was that ack-only message?
	if msgBatch.id == nil {
		// That was ack-only batch, nothing more to do with it.
		return NoMessages()
	}

	peerRecvAcksIn, _ := a.recvAcksIn.GetOrCreate(msgBatch.sender, func() map[int]*int { return make(map[int]*int) })

	batchAckedIn, exists := peerRecvAcksIn[*msgBatch.id]
	if exists {
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
		peerSentUnacked, exists := a.sentUnacked.Get(msgBatch.sender)
		if !exists {
			return NoMessages()
		}
		ackedBatch, exists := peerSentUnacked.Get(*batchAckedIn)
		if !exists {
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

	sender, _ := a.recvAcksIn.GetOrCreate(msgBatch.sender, func() map[int]*int { return make(map[int]*int) })
	sender[*msgBatch.id] = nil

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

	a.initPending.ForEach(func(nodeID NodeID, pending []Message) bool {
		if gr, ok := groupedMsgs[nodeID]; ok {
			groupedMsgs[nodeID] = append(gr, pending...)
		} else {
			groupedMsgs[nodeID] = pending
		}
		return true
	})
	a.initPending.Clear()

	batches := NoMessages()
	for nodeID, batchMsgs := range groupedMsgs {
		if initialized, exists := a.initialized.Get(nodeID); !exists || !initialized {
			pending, _ := a.initPending.GetOrCreate(nodeID, func() []Message { return make([]Message, 0, 1) })
			a.initPending.Set(nodeID, append(pending, batchMsgs...))
			batches.Add(&ackHandlerReset{BasicMessage: NewBasicMessage(nodeID), response: false, latestID: 0})
			continue
		}
		//
		// Assign batch ID.
		batchID, _ := a.counters.GetOrCreate(nodeID, func() int { return 0 })
		a.counters.Set(nodeID, batchID+1)

		//
		// Collect batches to be acknowledged and mark them as acknowledged.
		acks := []int{}
		if nodeRecvAcksIn, exists := a.recvAcksIn.Get(nodeID); exists {
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
		unackedMap, _ := a.sentUnacked.GetOrCreate(nodeID, func() *shrinkingmap.ShrinkingMap[int, *ackHandlerBatch] {
			return shrinkingmap.New[int, *ackHandlerBatch]()
		})
		unackedMap.Set(*batch.id, batch)
		batches.Add(batch)
	}
	return batches
}

////////////////////////////////////////////////////////////////////////////////
// ackHandlerReset

type ackHandlerReset struct {
	BasicMessage
	response bool `bcs:"export"`
	latestID int  `bcs:"export"`
}

var _ Message = new(ackHandlerReset)

func (msg *ackHandlerReset) MsgType() MessageType {
	return msgTypeAckHandlerReset
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

var _ Message = new(ackHandlerBatch)

func (msg *ackHandlerBatch) MsgType() MessageType {
	return msgTypeAckHandlerBatch
}

func (msg *ackHandlerBatch) Recipient() NodeID {
	return msg.recipient
}

func (msg *ackHandlerBatch) SetSender(sender NodeID) {
	msg.sender = sender
	for _, msg := range msg.msgs {
		msg.SetSender(sender)
	}
}

func (msg *ackHandlerBatch) MarshalBCS(e *bcs.Encoder) error {
	e.EncodeOptional(msg.id)

	msgsBytes, err := MarshalMessages(msg.msgs)
	if err != nil {
		return fmt.Errorf("msgs: %w", err)
	}

	e.Encode(msgsBytes)
	e.Encode(msg.acks)

	return nil
}

func (msg *ackHandlerBatch) UnmarshalBCS(d *bcs.Decoder) error {
	msg.id = nil

	d.DecodeOptional(&msg.id)

	msgsBytes := bcs.Decode[[][]byte](d)
	msg.msgs = make([]Message, len(msgsBytes))

	for i := range msgsBytes {
		var err error
		msg.msgs[i], err = msg.nestedGPA.UnmarshalMessage(msgsBytes[i])
		if err != nil {
			return err
		}
	}

	msg.acks = bcs.Decode[[]int](d)

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// ackHandlerTick

// Event representing a timer tick.
type ackHandlerTick struct {
	timestamp time.Time
}
