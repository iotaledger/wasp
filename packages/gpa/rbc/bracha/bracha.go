// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// package bracha implements Bracha's Reliable Broadcast, as described in
//
// 		Gabriel Bracha. 1987. Asynchronous byzantine agreement protocols. Inf. Comput.
// 		75, 2 (November 1, 1987), 130–143. DOI:https://doi.org/10.1016/0890-5401(87)90054-X
//
// Additionally a predicate is added as it was used in
//
// 		Sourav Das, Zhuolun Xiang, and Ling Ren. 2021. Asynchronous Data Dissemination
// 		and its Applications. In Proceedings of the 2021 ACM SIGSAC Conference on Computer
// 		and Communications Security (CCS '21). Association for Computing Machinery,
// 		New York, NY, USA, 2705–2721. DOI:https://doi.org/10.1145/3460120.3484808
//
// NOTE: Only a dedicated process can broadcast a value. Otherwise it would be a consensus.
package bracha

import (
	"fmt"
	"math"

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
	msgs := []gpa.Message{}
	for i := range r.peers {
		msgs = append(msgs, &msgBracha{
			t: msgBrachaTypeInitial,
			r: r.peers[i],
			v: inputVal,
		})
	}
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
	if r.echoSent {
		return []gpa.Message{}
	}
	if !r.predicate(msg.v) {
		r.pendingMsgs = append(r.pendingMsgs, msg)
		return []gpa.Message{}
	}
	msgs := []gpa.Message{}
	for i := range r.peers {
		msgs = append(msgs, &msgBracha{
			t: msgBrachaTypeEcho,
			r: r.peers[i],
			v: msg.v,
		})
	}
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
	// Support others, if quorum is enough.
	if count >= r.f+1 {
		return r.maybeSendEchoReady(msg.v)
	}
	return []gpa.Message{}
}

func (r *rbc) handlePredicateUpdate(msg msgPredicateUpdate) []gpa.Message {
	r.predicate = msg.predicate
	resendMsgs := r.pendingMsgs
	r.pendingMsgs = []gpa.Message{}
	return resendMsgs // The OwnHandler will resend them back.
}

func (r *rbc) maybeSendEchoReady(v []byte) []gpa.Message {
	msgs := []gpa.Message{}
	if !r.echoSent {
		for i := range r.peers {
			msgs = append(msgs, &msgBracha{
				t: msgBrachaTypeEcho,
				r: r.peers[i],
				v: v,
			})
		}
		r.echoSent = true
	}
	if !r.readySent {
		for i := range r.peers {
			msgs = append(msgs, &msgBracha{
				t: msgBrachaTypeReady,
				r: r.peers[i],
				v: v,
			})
		}
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

func (r *rbc) ensureValueStored(val []byte) hashing.HashValue {
	h := hashing.HashData(val)
	if _, ok := r.values[h]; ok {
		return h
	}
	r.values[h] = val
	return h
}
