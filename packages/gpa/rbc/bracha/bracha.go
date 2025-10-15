// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package bracha implements Bracha's Reliable Broadcast.
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
	"errors"
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/hashing"
)

type RBC struct {
	n           int
	f           int
	me          gpa.NodeID
	broadcaster gpa.NodeID
	maxMsgSize  int
	peers       []gpa.NodeID
	predicate   func([]byte) bool
	proposeSent bool
	msgRecv     map[gpa.NodeID]map[msgBrachaType]bool     // For tracking, who's messages are received.
	echoSent    bool                                      // Have we sent the ECHO messages?
	echoRecv    map[hashing.HashValue]map[gpa.NodeID]bool // Quorum counter for the ECHO messages.
	readySent   bool                                      // Have we sent the READY messages?
	readyRecv   map[hashing.HashValue]map[gpa.NodeID]bool // Quorum counter for the READY messages.
	output      []byte
	log         gpa.Logger
}

var _ gpa.GPA = &RBC{}

// New creates new instance of the RBC.
func New(peers []gpa.NodeID, f int, me, broadcaster gpa.NodeID, maxMsgSize int, predicate func([]byte) bool, log gpa.Logger) *gpa.OwnHandler[*RBC] {
	r := &RBC{
		n:           len(peers),
		f:           f,
		me:          me,
		broadcaster: broadcaster,
		maxMsgSize:  maxMsgSize,
		peers:       peers,
		predicate:   predicate,
		msgRecv:     map[gpa.NodeID]map[msgBrachaType]bool{},
		echoSent:    false,
		echoRecv:    make(map[hashing.HashValue]map[gpa.NodeID]bool),
		readySent:   false,
		readyRecv:   make(map[hashing.HashValue]map[gpa.NodeID]bool),
		output:      nil,
		log:         log,
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
func (r *RBC) Input(input gpa.Input) gpa.OutMessages {
	if r.broadcaster != r.me {
		panic(errors.New("only broadcaster is allowed to take an input"))
	}
	if r.proposeSent {
		panic(errors.New("input can only be supplied once"))
	}
	inputVal := input.([]byte)
	msgs := r.sendToAll(msgBrachaTypePropose, inputVal)
	r.proposeSent = true
	return msgs
}

// Implements the GPA interface.
func (r *RBC) Message(msg gpa.Message) gpa.OutMessages {
	switch msgT := msg.(type) {
	case *msgBracha:
		if !r.checkMsgRecv(msgT) {
			return nil
		}
		switch msgT.brachaType {
		case msgBrachaTypePropose:
			return r.handlePropose(msgT)
		case msgBrachaTypeEcho:
			return r.handleEcho(msgT)
		case msgBrachaTypeReady:
			return r.handleReady(msgT)
		default:
			r.log.LogWarnf("unexpected brachaType=%v in message: %+v", msgT.brachaType, msgT)
			return nil
		}
	default:
		panic(fmt.Errorf("unexpected message: %+v", msg))
	}
}

// Handle the PROPOSE messages.
//
//	06: upon receiving ⟨PROPOSE, 𝑀⟩ from the broadcaster do
//	07:     if 𝑃(𝑀) then
//	08:         send ⟨ECHO, 𝑀⟩ to all
func (r *RBC) handlePropose(msg *msgBracha) gpa.OutMessages {
	if msg.Sender() != r.broadcaster {
		// PROPOSE messages can only be sent by the broadcaster process.
		// Ignore all the rest.
		return nil
	}
	if !r.predicate(msg.value) {
		return nil
	}
	msgs := r.sendToAll(msgBrachaTypeEcho, msg.value)
	r.echoSent = true
	return msgs
}

// Handle the ECHO messages.
//
//	09: upon receiving 2𝑡 + 1 ⟨ECHO, 𝑀⟩ messages and not having sent a READY message do
//	10:     send ⟨READY, 𝑀⟩ to all
func (r *RBC) handleEcho(msg *msgBracha) gpa.OutMessages {
	//
	// Mark the message as received.
	h := r.valueHash(msg)
	r.markEchoRecv(h, msg)
	//
	// Send the READY message, if Byzantine quorum ⌈(n+f+1)/2⌉ of received ECHO messages is reached.
	// As there are only n distinct peers, every two Byzantine quorums overlap in at least one correct peer.
	// |echoRecv| ≥ ⌈(n+f+1)/2⌉ ⟺ |echoRecv| > ⌊(n+f)/2⌋
	if len(r.echoRecv[h]) > (r.n+r.f)/2 {
		return r.maybeSendReady(msg.value)
	}
	return nil
}

// Handle the READY messages.
//
//	11: upon receiving 𝑡 + 1 ⟨READY, 𝑀⟩ messages and not having sent a READY message do
//	12:     send ⟨READY, 𝑀⟩ to all
//	13: upon receiving 2𝑡 + 1 ⟨READY, 𝑀⟩ messages do
//	14:     output 𝑀
func (r *RBC) handleReady(msg *msgBracha) gpa.OutMessages {
	//
	// Mark the message as received.
	h := r.valueHash(msg)
	r.markReadyRecv(h, msg)
	count := len(r.readyRecv[h])
	//
	// Decide, if quorum is enough.
	if count > 2*r.f && r.output == nil {
		r.output = msg.value
	}
	//
	// Send the READY message, when a READY message was received from at least one honest peer.
	// This amplification assures totality.
	if count > r.f {
		return r.maybeSendReady(msg.value)
	}
	return nil
}

func (r *RBC) checkMsgRecv(msg *msgBracha) bool {
	if msg.value == nil || len(msg.value) > r.maxMsgSize {
		return false // Value not set, or is to big.
	}
	if mt, ok := r.msgRecv[msg.Sender()]; ok {
		if _, ok := mt[msg.brachaType]; !ok {
			mt[msg.brachaType] = true
			return true // OK, that was the first such message.
		}
		return false // Was already received before, ignore it.
	}
	return false // Unknown peer has sent it.
}

func (r *RBC) markEchoRecv(h hashing.HashValue, msg *msgBracha) {
	if _, ok := r.echoRecv[h]; !ok {
		r.echoRecv[h] = map[gpa.NodeID]bool{}
	}
	r.echoRecv[h][msg.Sender()] = true
}

func (r *RBC) markReadyRecv(h hashing.HashValue, msg *msgBracha) {
	if _, ok := r.readyRecv[h]; !ok {
		r.readyRecv[h] = map[gpa.NodeID]bool{}
	}
	r.readyRecv[h][msg.Sender()] = true
}

func (r *RBC) maybeSendReady(v []byte) gpa.OutMessages {
	if r.readySent {
		return nil
	}
	msgs := r.sendToAll(msgBrachaTypeReady, v)
	r.readySent = true
	return msgs
}

func (r *RBC) sendToAll(brachaType msgBrachaType, value []byte) gpa.OutMessages {
	msgs := make([]gpa.Message, len(r.peers))
	for i := range r.peers {
		msgs[i] = &msgBracha{
			BasicMessage: gpa.NewBasicMessage(r.peers[i]),
			brachaType:   brachaType,
			value:        value,
		}
	}
	return gpa.NoMessages().AddMany(msgs)
}

func (r *RBC) valueHash(msg *msgBracha) hashing.HashValue {
	return hashing.HashData(msg.value)
}

// Implements the GPA interface.
func (r *RBC) Output() gpa.Output {
	if r.output == nil {
		return nil // Return untyped nil!
	}
	return r.output
}

// Implements the GPA interface.
func (r *RBC) StatusString() string {
	return fmt.Sprintf(
		"{RBC:Bracha, n=%v, f=%v, output=%v,\nechoSent=%v, echoRecv=%v,\nreadySent=%v, readyRecv=%v}",
		r.n, r.f, r.output != nil, r.echoSent, r.echoRecv, r.readySent, r.readyRecv,
	)
}

// Implements the GPA interface.
func (r *RBC) UnmarshalMessage(data []byte) (gpa.Message, error) {
	return gpa.UnmarshalMessage(data, gpa.Mapper{
		msgType: func() gpa.Message { return new(msgBracha) },
	})
}
