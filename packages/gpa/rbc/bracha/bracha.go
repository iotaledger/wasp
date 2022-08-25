// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// package bracha implements Bracha's Reliable Broadcast, as described in
//
//      Gabriel Bracha. 1987. Asynchronous byzantine agreement protocols. Inf. Comput.
//      75, 2 (November 1, 1987), 130–143. DOI:https://doi.org/10.1016/0890-5401(87)90054-X
//
// The following pseudo-code is taken from "FIG. 1. The broadcast primitive" in the above
// paper. Here the paper assumes $0 ≤ t < n/3$.
//
//      Broadcast(v)
//      step 0. (By process p)
//          Send (initial, v) to all the processes.
//      step 1. Wait until the receipt of,
//              one (initial, v) message
//              or (n+t)/2 (echo, v) messages
//              or (t+l) (ready, v) messages
//              for some v.
//          Send (echo, v) to all the processes.
//      step 2. Wait until the receipt of,
//              (n+t)/2 (echo, v) messages
//              or (t+1) (ready, v) messages
//              (including messages received in step 1)
//              for some v.
//          Send (ready, v) to all the processes.
//      step 3. Wait until the receipt of,
//              2t+1 (ready, v) messages
//              (including messages received in step 1 and step 2) for some v.
//          Accept v.
//
// Additionally a predicate is added as it was used in
//
//      Sourav Das, Zhuolun Xiang, and Ling Ren. 2021. Asynchronous Data Dissemination
//      and its Applications. In Proceedings of the 2021 ACM SIGSAC Conference on Computer
//      and Communications Security (CCS '21). Association for Computing Machinery,
//      New York, NY, USA, 2705–2721. DOI:https://doi.org/10.1145/3460120.3484808
//
// NOTE: Only a dedicated process can broadcast a value. Otherwise it would be a consensus.
package bracha

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/hashing"
	"golang.org/x/xerrors"
)

type rbc struct {
	n           int
	f           int
	me          gpa.NodeID
	broadcaster gpa.NodeID
	peers       []gpa.NodeID
	predicate   func([]byte) bool
	pendingMsgs []gpa.Message // Messages that don't satisfy the predicate.
	initialSent bool
	values      map[hashing.HashValue][]byte              // Map hashes to actual values.
	echoSent    bool                                      // Have we sent the ECHO messages?
	echoRecv    map[hashing.HashValue]map[gpa.NodeID]bool // Quorum counter for the ECHO messages.
	readySent   bool                                      // Have we sent the READY messages?
	readyRecv   map[hashing.HashValue]map[gpa.NodeID]bool // Quorum counter for the READY messages.
	output      hashing.HashValue
}

var _ gpa.GPA = &rbc{}

// Update predicate for the RBC instance.
func SendPredicateUpdate(rbc gpa.GPA, me gpa.NodeID, predicate func([]byte) bool) []gpa.Message {
	return rbc.Message(MakePredicateUpdateMsg(me, predicate))
}

func MakePredicateUpdateMsg(me gpa.NodeID, predicate func([]byte) bool) gpa.Message {
	return &msgPredicateUpdate{me: me, predicate: predicate}
}

func New(peers []gpa.NodeID, f int, me, broadcaster gpa.NodeID, predicate func([]byte) bool) gpa.GPA {
	r := &rbc{
		n:           len(peers),
		f:           f,
		me:          me,
		broadcaster: broadcaster,
		peers:       peers,
		predicate:   predicate,
		pendingMsgs: []gpa.Message{},
		values:      make(map[hashing.HashValue][]byte),
		echoSent:    false,
		echoRecv:    make(map[hashing.HashValue]map[gpa.NodeID]bool),
		readySent:   false,
		readyRecv:   make(map[hashing.HashValue]map[gpa.NodeID]bool),
		output:      hashing.NilHash,
	}
	return gpa.NewOwnHandler(me, r)
}

func (r *rbc) Input(input gpa.Input) []gpa.Message {
	if r.broadcaster != r.me {
		panic(xerrors.Errorf("only broadcaster is allowed to take an input"))
	}
	if r.initialSent {
		panic(xerrors.Errorf("input can only be supplied once"))
	}
	inputVal := input.([]byte)
	r.ensureValueStored(inputVal)
	msgs := r.sendToAll(msgBrachaTypeInitial, inputVal)
	r.initialSent = true
	return msgs
}

func (r *rbc) Message(msg gpa.Message) []gpa.Message {
	switch msgT := msg.(type) {
	case *msgBracha:
		switch msgT.t {
		case msgBrachaTypeInitial:
			return r.handleInitial(msgT)
		case msgBrachaTypeEcho:
			return r.handleEcho(msgT)
		case msgBrachaTypeReady:
			return r.handleReady(msgT)
		default:
			panic(xerrors.Errorf("unexpected message: %+v", msgT))
		}
	case *msgPredicateUpdate:
		return r.handlePredicateUpdate(*msgT)
	default:
		panic(xerrors.Errorf("unexpected message: %+v", msg))
	}
}

func (r *rbc) handleInitial(msg *msgBracha) []gpa.Message {
	if msg.s != r.broadcaster {
		// Initial messages can only be sent by the broadcaster process.
		// Ignore all the rest.
		return gpa.NoMessages()
	}
	if r.echoSent {
		return gpa.NoMessages()
	}
	if !r.predicate(msg.v) {
		r.pendingMsgs = append(r.pendingMsgs, msg)
		return gpa.NoMessages()
	}
	msgs := r.sendToAll(msgBrachaTypeEcho, msg.v)
	r.echoSent = true
	return msgs
}

func (r *rbc) handleEcho(msg *msgBracha) []gpa.Message {
	//
	// Mark the message as received.
	h := r.ensureValueStored(msg.v)
	if _, ok := r.echoRecv[h]; !ok {
		r.echoRecv[h] = map[gpa.NodeID]bool{}
	}
	r.echoRecv[h][msg.s] = true
	//
	// Send the READY message, if Byzantine quorum ⌈(n+f+1)/2⌉ of received ECHO messages is reached.
	// As there are only n distinct peers, every two Byzantine quorums overlap in at least one correct peer.
	// |echoRecv| ≥ ⌈(n+f+1)/2⌉ ⟺ |echoRecv| > ⌊(n+f)/2⌋
	if len(r.echoRecv[h]) > (r.n+r.f)/2 {
		return r.maybeSendEchoReady(msg.v)
	}
	return []gpa.Message{}
}

func (r *rbc) handleReady(msg *msgBracha) []gpa.Message {
	//
	// Mark the message as received.
	h := r.ensureValueStored(msg.v)
	if _, ok := r.readyRecv[h]; !ok {
		r.readyRecv[h] = map[gpa.NodeID]bool{}
	}
	r.readyRecv[h][msg.s] = true
	count := len(r.readyRecv[h])
	//
	// Decide, if quorum is enough.
	if count >= 2*r.f+1 && r.output == hashing.NilHash {
		r.output = h
	}
	//
	// Send the READY message, when a READY message was received from at least one honest peer.
	// This amplification assures totality.
	if count > r.f {
		return r.maybeSendEchoReady(msg.v)
	}
	return []gpa.Message{}
}

func (r *rbc) handlePredicateUpdate(msg msgPredicateUpdate) []gpa.Message {
	r.predicate = msg.predicate
	resendMsgs := r.pendingMsgs
	r.pendingMsgs = []gpa.Message{}
	//
	// Resend the messages again.
	// The OwnHandler overrides the sender, thus we can't rely on it here.
	msgs := gpa.NoMessages()
	for i := range resendMsgs {
		msgs = append(msgs, r.Message(resendMsgs[i])...)
	}
	return msgs
}

func (r *rbc) maybeSendEchoReady(v []byte) []gpa.Message {
	msgs := []gpa.Message{}
	if !r.echoSent {
		msgs = append(msgs, r.sendToAll(msgBrachaTypeEcho, v)...)
		r.echoSent = true
	}
	if !r.readySent {
		msgs = append(msgs, r.sendToAll(msgBrachaTypeReady, v)...)
		r.readySent = true
	}
	return msgs
}

func (r *rbc) Output() gpa.Output {
	if r.output == hashing.NilHash {
		return nil
	}
	return r.values[r.output]
}

func (r *rbc) StatusString() string {
	return fmt.Sprintf(
		"{RBC:Bracha, n=%v, f=%v, output=%v,\nechoSent=%v, echoRecv=%v,\nreadySent=%v, readyRecv=%v}",
		r.n, r.f, r.output, r.echoSent, r.echoRecv, r.readySent, r.readyRecv,
	)
}

func (r *rbc) UnmarshalMessage(data []byte) (gpa.Message, error) {
	m := &msgBracha{}
	if err := m.UnmarshalBinary(data); err != nil {
		return nil, xerrors.Errorf("cannot unmarshal RBC:msgBracha message: %w", err)
	}
	return m, nil
}

func (r *rbc) sendToAll(t msgBrachaType, v []byte) []gpa.Message {
	msgs := make([]gpa.Message, len(r.peers))
	for i := range r.peers {
		msgs[i] = &msgBracha{
			t: t,
			r: r.peers[i],
			v: v,
		}
	}
	return msgs
}

func (r *rbc) ensureValueStored(val []byte) hashing.HashValue {
	h := hashing.HashData(val)
	if _, ok := r.values[h]; ok {
		return h
	}
	r.values[h] = val
	return h
}
