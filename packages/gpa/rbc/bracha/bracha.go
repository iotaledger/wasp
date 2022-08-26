// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// package bracha implements Bracha's Reliable Broadcast.
// The original version of this RBC can be found here (see "FIG. 1. The broadcast primitive"):
//
//	Gabriel Bracha. 1987. Asynchronous byzantine agreement protocols. Inf. Comput.
//	75, 2 (November 1, 1987), 130–143. DOI:https://doi.org/10.1016/0890-5401(87)90054-X
//
// Here we follow the algorithm presentation from (see "Algorithm 2 Bracha’s RBC [14]"):
//
//	Sourav Das, Zhuolun Xiang, and Ling Ren. 2021. Asynchronous Data Dissemination
//	and its Applications. In Proceedings of the 2021 ACM SIGSAC Conference on Computer
//	and Communications Security (CCS '21). Association for Computing Machinery,
//	New York, NY, USA, 2705–2721. DOI:https://doi.org/10.1145/3460120.3484808
//
// The algorithms differs a bit. The latter supports predicates and also it don't
// imply sending ECHO messages upon receiving F+1 READY messages. The pseudo-code
// from the Das et al.:
//
//	01: // only broadcaster node
//	02: input 𝑀
//	03: send ⟨PROPOSE, 𝑀⟩ to all
//	04: // all nodes
//	05: input 𝑃(·) // predicate 𝑃(·) returns true unless otherwise specified.
//	06: upon receiving ⟨PROPOSE, 𝑀⟩ from the broadcaster do
//	07:     if 𝑃(𝑀) then
//	08:         send ⟨ECHO, 𝑀⟩ to all
//	09: upon receiving 2𝑡 + 1 ⟨ECHO, 𝑀⟩ messages and not having sent a READY message do
//	10:     send ⟨READY, 𝑀⟩ to all
//	11: upon receiving 𝑡 + 1 ⟨READY, 𝑀⟩ messages and not having sent a READY message do
//	12:     send ⟨READY, 𝑀⟩ to all
//	13: upon receiving 2𝑡 + 1 ⟨READY, 𝑀⟩ messages do
//	14:     output 𝑀
//
// In the above 𝑡 is "Given a network of 𝑛 nodes, of which up to 𝑡 could be malicious",
// thus that's the parameter F in the specification bellow.
//
// On the predicates. If they are updated via `MakePredicateUpdateMsg` and similar,
// they have to be monotonic. I.e. if a predicate was true for the broadcaster's
// message, then all the following predicates supplied to the algorithm must be
// true for that message as well.
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
	maxMsgSize  int
	peers       []gpa.NodeID
	predicate   func([]byte) bool
	pendingPMsg *msgBracha // PROPOSE message received, but not satisfying a predicate, if any.
	proposeSent bool
	msgRecv     map[gpa.NodeID]map[msgBrachaType]bool     // For tracking, who's messages are received.
	echoSent    bool                                      // Have we sent the ECHO messages?
	echoRecv    map[hashing.HashValue]map[gpa.NodeID]bool // Quorum counter for the ECHO messages.
	readySent   bool                                      // Have we sent the READY messages?
	readyRecv   map[hashing.HashValue]map[gpa.NodeID]bool // Quorum counter for the READY messages.
	output      []byte
}

var _ gpa.GPA = &rbc{}

// Update predicate for the RBC instance.
func SendPredicateUpdate(rbc gpa.GPA, me gpa.NodeID, predicate func([]byte) bool) []gpa.Message {
	return rbc.Message(MakePredicateUpdateMsg(me, predicate))
}

// Create a message for sending it later.
func MakePredicateUpdateMsg(me gpa.NodeID, predicate func([]byte) bool) gpa.Message {
	return &msgPredicateUpdate{me: me, predicate: predicate}
}

// Create new instance of the RBC.
func New(peers []gpa.NodeID, f int, me, broadcaster gpa.NodeID, maxMsgSize int, predicate func([]byte) bool) gpa.GPA {
	r := &rbc{
		n:           len(peers),
		f:           f,
		me:          me,
		broadcaster: broadcaster,
		maxMsgSize:  maxMsgSize,
		peers:       peers,
		predicate:   predicate,
		pendingPMsg: nil,
		msgRecv:     map[gpa.NodeID]map[msgBrachaType]bool{},
		echoSent:    false,
		echoRecv:    make(map[hashing.HashValue]map[gpa.NodeID]bool),
		readySent:   false,
		readyRecv:   make(map[hashing.HashValue]map[gpa.NodeID]bool),
		output:      nil,
	}
	for i := range peers {
		r.msgRecv[peers[i]] = map[msgBrachaType]bool{}
	}
	return gpa.NewOwnHandler(me, r)
}

// Implements the GPA interface.
//
//	01: // only broadcaster node
//	02: input 𝑀
//	03: send ⟨PROPOSE, 𝑀⟩ to all
func (r *rbc) Input(input gpa.Input) []gpa.Message {
	if r.broadcaster != r.me {
		panic(xerrors.Errorf("only broadcaster is allowed to take an input"))
	}
	if r.proposeSent {
		panic(xerrors.Errorf("input can only be supplied once"))
	}
	inputVal := input.([]byte)
	msgs := r.sendToAll(msgBrachaTypePropose, inputVal)
	r.proposeSent = true
	return msgs
}

// Implements the GPA interface.
func (r *rbc) Message(msg gpa.Message) []gpa.Message {
	switch msgT := msg.(type) {
	case *msgBracha:
		if !r.checkMsgRecv(msgT) {
			return gpa.NoMessages()
		}
		switch msgT.t {
		case msgBrachaTypePropose:
			return r.handlePropose(msgT)
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

// Handle the PROPOSE messages.
//
//	06: upon receiving ⟨PROPOSE, 𝑀⟩ from the broadcaster do
//	07:     if 𝑃(𝑀) then
//	08:         send ⟨ECHO, 𝑀⟩ to all
func (r *rbc) handlePropose(msg *msgBracha) []gpa.Message {
	if msg.s != r.broadcaster {
		// PROPOSE messages can only be sent by the broadcaster process.
		// Ignore all the rest.
		return gpa.NoMessages()
	}
	if r.echoSent || r.pendingPMsg != nil {
		// PROPOSE message was already received, ignore this one.
		return gpa.NoMessages()
	}
	if !r.predicate(msg.v) {
		r.pendingPMsg = msg
		return gpa.NoMessages()
	}
	msgs := r.sendToAll(msgBrachaTypeEcho, msg.v)
	r.echoSent = true
	return msgs
}

// Handle the ECHO messages.
//
//	09: upon receiving 2𝑡 + 1 ⟨ECHO, 𝑀⟩ messages and not having sent a READY message do
//	10:     send ⟨READY, 𝑀⟩ to all
func (r *rbc) handleEcho(msg *msgBracha) []gpa.Message {
	//
	// Mark the message as received.
	h := r.valueHash(msg)
	r.markEchoRecv(h, msg)
	//
	// Send the READY message, if Byzantine quorum ⌈(n+f+1)/2⌉ of received ECHO messages is reached.
	// As there are only n distinct peers, every two Byzantine quorums overlap in at least one correct peer.
	if len(r.echoRecv[h]) > 2*r.f {
		return r.maybeSendReady(msg.v)
	}
	return gpa.NoMessages()
}

// Handle the READY messages.
//
//	11: upon receiving 𝑡 + 1 ⟨READY, 𝑀⟩ messages and not having sent a READY message do
//	12:     send ⟨READY, 𝑀⟩ to all
//	13: upon receiving 2𝑡 + 1 ⟨READY, 𝑀⟩ messages do
//	14:     output 𝑀
func (r *rbc) handleReady(msg *msgBracha) []gpa.Message {
	//
	// Mark the message as received.
	h := r.valueHash(msg)
	r.markReadyRecv(h, msg)
	count := len(r.readyRecv[h])
	//
	// Decide, if quorum is enough.
	if count > 2*r.f && r.output == nil {
		r.output = msg.v
	}
	//
	// Send the READY message, when a READY message was received from at least one honest peer.
	// This amplification assures totality.
	if count > r.f {
		return r.maybeSendReady(msg.v)
	}
	return gpa.NoMessages()
}

func (r *rbc) handlePredicateUpdate(msg msgPredicateUpdate) []gpa.Message {
	r.predicate = msg.predicate
	if r.pendingPMsg == nil {
		return gpa.NoMessages()
	}
	//
	// Try to process the PROPOSE message again, if it was postponed.
	proposeMsg := r.pendingPMsg
	r.pendingPMsg = nil
	return r.handlePropose(proposeMsg)
}

func (r *rbc) checkMsgRecv(msg *msgBracha) bool {
	if msg.v == nil || len(msg.v) > r.maxMsgSize {
		return false // Value not set, or is to big.
	}
	if mt, ok := r.msgRecv[msg.s]; ok {
		if _, ok := mt[msg.t]; !ok {
			mt[msg.t] = true
			return true // OK, that was the first such message.
		}
		return false // Was already received before, ignore it.
	}
	return false // Unknown peer has sent it.
}

func (r *rbc) markEchoRecv(h hashing.HashValue, msg *msgBracha) {
	if _, ok := r.echoRecv[h]; !ok {
		r.echoRecv[h] = map[gpa.NodeID]bool{}
	}
	r.echoRecv[h][msg.s] = true
}

func (r *rbc) markReadyRecv(h hashing.HashValue, msg *msgBracha) {
	if _, ok := r.readyRecv[h]; !ok {
		r.readyRecv[h] = map[gpa.NodeID]bool{}
	}
	r.readyRecv[h][msg.s] = true
}

func (r *rbc) maybeSendReady(v []byte) []gpa.Message {
	msgs := gpa.NoMessages()
	if !r.readySent {
		msgs = append(msgs, r.sendToAll(msgBrachaTypeReady, v)...)
		r.readySent = true
	}
	return msgs
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

func (r *rbc) valueHash(msg *msgBracha) hashing.HashValue {
	return hashing.HashData(msg.v)
}

// Implements the GPA interface.
func (r *rbc) Output() gpa.Output {
	return r.output
}

// Implements the GPA interface.
func (r *rbc) StatusString() string {
	return fmt.Sprintf(
		"{RBC:Bracha, n=%v, f=%v, output=%v,\nechoSent=%v, echoRecv=%v,\nreadySent=%v, readyRecv=%v}",
		r.n, r.f, r.output != nil, r.echoSent, r.echoRecv, r.readySent, r.readyRecv,
	)
}

// Implements the GPA interface.
func (r *rbc) UnmarshalMessage(data []byte) (gpa.Message, error) {
	m := &msgBracha{}
	if err := m.UnmarshalBinary(data); err != nil {
		return nil, xerrors.Errorf("cannot unmarshal RBC:msgBracha message: %w", err)
	}
	return m, nil
}
