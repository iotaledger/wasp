// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package gpa

import (
	"fmt"
	"slices"
	"time"

	"fortio.org/safecast"
	"github.com/samber/lo"

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
	initPending  *shrinkingmap.ShrinkingMap[NodeID, []MessagePayload]
	counters     *shrinkingmap.ShrinkingMap[NodeID, int] // For numbering the outgoing messages.
	sentUnacked  *shrinkingmap.ShrinkingMap[NodeID, *shrinkingmap.ShrinkingMap[int, *ackHandlerBatch]]
	recvAcksIn   *shrinkingmap.ShrinkingMap[NodeID, map[int]*int]
}

type AckHandler interface {
	GPA
	DismissPeer(peerID NodeID) // To avoid resending messages to dead peers.
	MakeTickInput(time.Time) Input
	NestedMessage(msg *MessageIn) []*MessageOut
	NestedCall(c func(GPA) []*MessageOut) []*MessageOut
}

var _ AckHandler = &ackHandler{}

func NewAckHandler(me NodeID, nested GPA, resendPeriod time.Duration) AckHandler {
	return &ackHandler{
		me:           me,
		nested:       nested,
		resendPeriod: resendPeriod,
		initialized:  shrinkingmap.New[NodeID, bool](),
		initPending:  shrinkingmap.New[NodeID, []MessagePayload](),
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

func (a *ackHandler) Input(input Input) []*MessageOut {
	switch input := input.(type) {
	case *ackHandlerTick:
		return a.handleTickMsg(input)
	default:
		return a.makeBatches(a.nested.Input(input))
	}
}

func (a *ackHandler) Message(msg *MessageIn) []*MessageOut {
	switch msg.Payload.(type) {
	case *ackHandlerReset:
		return a.handleResetMsg(AsTypedMessageIn[*ackHandlerReset](msg))
	case *ackHandlerBatch:
		return a.handleBatchMsg(AsTypedMessageIn[*ackHandlerBatch](msg))
	default:
		panic(fmt.Errorf("unexpected message type: %+v", msg))
	}
}

func (a *ackHandler) NestedMessage(msg *MessageIn) []*MessageOut {
	return a.makeBatches(a.nested.Message(msg))
}

func (a *ackHandler) NestedCall(c func(GPA) []*MessageOut) []*MessageOut {
	return a.makeBatches(c(a.nested))
}

func (a *ackHandler) Output() Output {
	return a.nested.Output()
}

func (a *ackHandler) StatusString() string {
	return fmt.Sprintf("{ACK:%s}", a.nested.StatusString())
}

func (a *ackHandler) UnmarshalPayload(data []byte) (MessagePayload, error) {
	msg, err := UnmarshalPayload(data, PayloadAllocator{
		msgTypeAckHandlerReset: func() MessagePayload { return &ackHandlerReset{} },
		msgTypeAckHandlerBatch: func() MessagePayload { return &ackHandlerBatch{nestedGPA: a.nested} },
	})
	if err != nil {
		fmt.Printf("ack, err=%v\n", err) // TODO: Clean this up.
	}
	return msg, err
}

func (a *ackHandler) handleTickMsg(msg *ackHandlerTick) []*MessageOut {
	resendOlderThan := msg.timestamp.Add(-a.resendPeriod)
	var resendMsgs []*MessageOut
	a.sentUnacked.ForEach(func(nodeID NodeID, nodeSentUnacked *shrinkingmap.ShrinkingMap[int, *ackHandlerBatch]) bool {
		nodeSentUnacked.ForEach(func(batchID int, batch *ackHandlerBatch) bool {
			if batch.sent == nil {
				// Don't resend, just mark the current timestamp.
				// We have sent it after the previous tick.
				batch.sent = &msg.timestamp
			} else if batch.sent.Before(resendOlderThan) {
				// Resend it, timeout is already passed.
				batch.sent = &msg.timestamp
				resendMsgs = append(resendMsgs, NewMessageOut(nodeID, batch))
			}
			return true
		})
		return true
	})

	a.initPending.ForEachKey(func(nodeID NodeID) bool {
		resendMsgs = append(resendMsgs, NewMessageOut(nodeID, &ackHandlerReset{
			response: false,
			latestID: 0,
		}))
		return true
	})
	return resendMsgs
}

func (a *ackHandler) handleResetMsg(msg *TypedMessageIn[*ackHandlerReset]) []*MessageOut {
	from := msg.Sender
	if !msg.Payload.response {
		maxID := 0
		if recvAcksIn, exists := a.recvAcksIn.Get(msg.Sender); exists {
			for id := range recvAcksIn {
				if id > maxID {
					maxID = id
				}
			}
		}
		return []*MessageOut{NewMessageOut(msg.Sender, &ackHandlerReset{
			response: true,
			latestID: maxID,
		})}
	}
	if ini, exists := a.initialized.Get(from); exists && ini {
		return nil
	}
	a.counters.Set(msg.Sender, msg.Payload.latestID+1)
	a.initialized.Set(msg.Sender, true)
	return a.makeBatches(nil)
}

func (a *ackHandler) handleBatchMsg(msgBatch *TypedMessageIn[*ackHandlerBatch]) []*MessageOut {
	//
	// Process the received acknowledgements.
	// Drop all the outgoing batches, that are now acknowledged.
	for _, ackedBatchID := range msgBatch.Payload.acks {
		if unacked, exists := a.sentUnacked.Get(msgBatch.Sender); exists {
			unacked.Delete(ackedBatchID)
		}
	}
	//
	// Was that ack-only message?
	if msgBatch.Payload.id == nil {
		// That was ack-only batch, nothing more to do with it.
		return nil
	}

	peerRecvAcksIn, _ := a.recvAcksIn.GetOrCreate(msgBatch.Sender, func() map[int]*int { return make(map[int]*int) })

	batchAckedIn, exists := peerRecvAcksIn[*msgBatch.Payload.id]
	if exists {
		// Was received already before.
		if batchAckedIn == nil {
			// Not acknowledged yet, just send an ack-only message for now.
			// The sender has already re-sent the message, so it waits for the ack.
			return []*MessageOut{NewMessageOut(msgBatch.Sender, &ackHandlerBatch{
				id:   nil,                         // That's ack-only.
				msgs: nil,                         // No payload.
				acks: []int{*msgBatch.Payload.id}, // Ack single message.
				sent: nil,                         // We will not track this message, it has no payload.
			})}
		}
		//
		// We have acked it already. If we have the batch with an ack, we
		// resent it. Otherwise the ack was already acked and this message
		// is outdated and can be ignored.
		peerSentUnacked, exists := a.sentUnacked.Get(msgBatch.Sender)
		if !exists {
			return nil
		}
		ackedBatch, exists := peerSentUnacked.Get(*batchAckedIn)
		if !exists {
			return nil
		}
		now := time.Now()
		ackedBatch.sent = &now
		return []*MessageOut{NewMessageOut(msgBatch.Sender, ackedBatch)}
	}
	//
	// That's a new batch, we have to process it.
	var nestedMsgs []*MessageOut
	for _, p := range msgBatch.Payload.msgs {
		nestedMsgs = slices.Concat(nestedMsgs, a.nested.Message(NewMessageIn(msgBatch.Sender, p)))
	}

	sender, _ := a.recvAcksIn.GetOrCreate(msgBatch.Sender, func() map[int]*int { return make(map[int]*int) })
	sender[*msgBatch.Payload.id] = nil

	return a.makeBatches(nestedMsgs)
}

func (a *ackHandler) makeBatches(msgs []*MessageOut) []*MessageOut {
	groupedMsgs := lo.MapEntries(
		lo.GroupBy(msgs, func(msg *MessageOut) NodeID { return msg.Recipient }),
		func(nodeID NodeID, msgsForNode []*MessageOut) (NodeID, []MessagePayload) {
			return nodeID, lo.Map(msgsForNode, func(msg *MessageOut, _ int) MessagePayload {
				return msg.Payload
			})
		},
	)

	a.initPending.ForEach(func(nodeID NodeID, pending []MessagePayload) bool {
		if gr, ok := groupedMsgs[nodeID]; ok {
			groupedMsgs[nodeID] = append(gr, pending...)
		} else {
			groupedMsgs[nodeID] = pending
		}
		return true
	})
	a.initPending.Clear()

	var batches []*MessageOut
	for nodeID, batchMsgs := range groupedMsgs {
		if initialized, exists := a.initialized.Get(nodeID); !exists || !initialized {
			pending, _ := a.initPending.GetOrCreate(nodeID, func() []MessagePayload { return make([]MessagePayload, 0, 1) })
			a.initPending.Set(nodeID, append(pending, batchMsgs...))
			batches = append(batches, NewMessageOut(nodeID, &ackHandlerReset{
				response: false,
				latestID: 0,
			}))
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
			id:   &batchID,
			acks: acks,
			msgs: batchMsgs,
			sent: nil, // Will be set after first resend, to avoid resend too early.
		}
		unackedMap, _ := a.sentUnacked.GetOrCreate(nodeID, func() *shrinkingmap.ShrinkingMap[int, *ackHandlerBatch] {
			return shrinkingmap.New[int, *ackHandlerBatch]()
		})
		unackedMap.Set(*batch.id, batch)
		batches = append(batches, NewMessageOut(nodeID, batch))
	}
	return batches
}

////////////////////////////////////////////////////////////////////////////////
// ackHandlerReset

type ackHandlerReset struct {
	response bool `bcs:"export"`
	latestID int  `bcs:"export"`
}

var _ MessagePayload = new(ackHandlerReset)

func (msg *ackHandlerReset) MsgType() MessageType {
	return msgTypeAckHandlerReset
}

////////////////////////////////////////////////////////////////////////////////
// ackHandlerBatch

// Message conveying the message batches and acknowledgements.
type ackHandlerBatch struct {
	id        *int             // That's ACK only, if nil.
	msgs      []MessagePayload // Messages in the batch.
	acks      []int            // Acknowledged batches.
	sent      *time.Time       // Transient, only used for outgoing messages, not sent to the outside.
	nestedGPA GPA              // Transient, for un-marshaling only.
}

var _ MessagePayload = new(ackHandlerBatch)

func (msg *ackHandlerBatch) MsgType() MessageType {
	return msgTypeAckHandlerBatch
}

func (msg *ackHandlerBatch) MarshalBCS(e *bcs.Encoder) error {
	e.EncodeOptional(msg.id)

	n, err := safecast.Convert[uint16](len(msg.msgs))
	if err != nil {
		return fmt.Errorf("too many nested messages to marshal: %w", err)
	}
	e.Encode(n)
	for _, p := range msg.msgs {
		msgBytes, err := MarshalPayload(p)
		if err != nil {
			return fmt.Errorf("marshaling nested payload: %w", err)
		}
		e.Encode(msgBytes)
	}

	e.Encode(msg.acks)
	return nil
}

func (msg *ackHandlerBatch) UnmarshalBCS(d *bcs.Decoder) error {
	msg.id = nil
	d.DecodeOptional(&msg.id)

	var n uint16
	d.Decode(&n)
	msg.msgs = make([]MessagePayload, n)
	for i := uint16(0); i < n; i++ {
		msgBytes := bcs.Decode[[]byte](d)
		payload, err := msg.nestedGPA.UnmarshalPayload(msgBytes)
		if err != nil {
			return fmt.Errorf("msgs[%d]: %w", i, err)
		}
		msg.msgs[i] = payload
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
