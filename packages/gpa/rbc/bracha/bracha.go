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
	"math"

	"github.com/iotaledger/hive.go/serializer/v2"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/hashing"
	"golang.org/x/xerrors"
)

type RBC struct {
	n           int
	f           int
	me          gpa.NodeID
	broadcaster gpa.NodeID
	peers       []gpa.NodeID
	predicate   func([]byte) bool
	initialSent bool
	values      map[hashing.HashValue][]byte              // Map hashes to actual values.
	echoSent    bool                                      // Have we sent the ECHO messages?
	echoRecv    map[hashing.HashValue]map[gpa.NodeID]bool // Quorum counter for the ECHO messages.
	readySent   bool                                      // Have we sent the READY messages?
	readyRecv   map[hashing.HashValue]map[gpa.NodeID]bool // Quorum counter for the READY messages.
	output      hashing.HashValue
}

func New(peers []gpa.NodeID, f int, me, broadcaster gpa.NodeID, predicate func([]byte) bool) gpa.GPA {
	r := &RBC{
		n:           len(peers),
		f:           f,
		me:          me,
		broadcaster: broadcaster,
		peers:       peers,
		predicate:   predicate,
		values:      make(map[hashing.HashValue][]byte),
		echoSent:    false,
		echoRecv:    make(map[hashing.HashValue]map[gpa.NodeID]bool),
		readySent:   false,
		readyRecv:   make(map[hashing.HashValue]map[gpa.NodeID]bool),
		output:      hashing.NilHash,
	}
	return gpa.NewOwnHandler(me, r)
}

func (r *RBC) Input(input gpa.Input) []gpa.Message {
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
		msgs = append(msgs, &message{
			t: msgInitial,
			s: r.me,
			r: r.peers[i],
			v: inputVal,
		})
	}
	r.initialSent = true
	return msgs
}

func (r *RBC) Message(msg gpa.Message) []gpa.Message {
	m := msg.(*message)
	switch m.t {
	case msgInitial:
		return r.handleInitial(m)
	case msgEcho:
		return r.handleEcho(m)
	case msgReady:
		return r.handleReady(m)
	default:
		panic(xerrors.Errorf("unexpected message: %+v", m))
	}
}

func (r *RBC) handleInitial(msg *message) []gpa.Message {
	if r.echoSent || !r.predicate(msg.v) {
		return []gpa.Message{}
	}
	msgs := []gpa.Message{}
	for i := range r.peers {
		msgs = append(msgs, &message{
			t: msgEcho,
			s: r.me,
			r: r.peers[i],
			v: msg.v,
		})
	}
	r.echoSent = true
	return msgs
}

func (r *RBC) handleEcho(msg *message) []gpa.Message {
	//
	// Mark the message as received.
	h := r.ensureValueStored(msg.v)
	if _, ok := r.echoRecv[h]; !ok {
		r.echoRecv[h] = map[gpa.NodeID]bool{}
	}
	r.echoRecv[h][msg.s] = true
	//
	// Send the READY messages and support other ECHO messages, if quorum is reached.
	if float64(len(r.echoRecv[h])) >= math.Ceil(float64(r.n+r.f)/2.0) {
		return r.maybeSendEchoReady(msg.v)
	}
	return []gpa.Message{}
}

func (r *RBC) handleReady(msg *message) []gpa.Message {
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

func (r *RBC) maybeSendEchoReady(v []byte) []gpa.Message {
	msgs := []gpa.Message{}
	if !r.echoSent {
		for i := range r.peers {
			msgs = append(msgs, &message{
				t: msgEcho,
				s: r.me,
				r: r.peers[i],
				v: v,
			})
		}
		r.echoSent = true
	}
	if !r.readySent {
		for i := range r.peers {
			msgs = append(msgs, &message{
				t: msgReady,
				s: r.me,
				r: r.peers[i],
				v: v,
			})
		}
		r.readySent = true
	}
	return msgs
}

func (r *RBC) Output() gpa.Output {
	if r.output == hashing.NilHash {
		return nil
	}
	return r.values[r.output]
}

func (r *RBC) ensureValueStored(val []byte) hashing.HashValue {
	h := hashing.HashData(val)
	if _, ok := r.values[h]; ok {
		return h
	}
	r.values[h] = val
	return h
}

//////////////////// message ////////////////////////////////////////////////

const (
	msgInitial byte = iota
	msgEcho
	msgReady
)

type message struct {
	t byte       // Type
	s gpa.NodeID // Sender
	r gpa.NodeID // Recipient
	v []byte     // Value
}

var _ gpa.Message = &message{}

func (m *message) Recipient() gpa.NodeID {
	return m.r
}

func (m *message) MarshalBinary() ([]byte, error) {
	return serializer.NewSerializer().
		WriteByte(m.t, func(err error) error { return xerrors.Errorf("unable to serialize t: %w", err) }).
		WriteString(string(m.s), serializer.SeriLengthPrefixTypeAsUint16, func(err error) error { return xerrors.Errorf("unable to serialize s: %w", err) }).
		WriteString(string(m.r), serializer.SeriLengthPrefixTypeAsUint16, func(err error) error { return xerrors.Errorf("unable to serialize r: %w", err) }).
		WriteVariableByteSlice(m.v, serializer.SeriLengthPrefixTypeAsUint16, func(err error) error { return xerrors.Errorf("unable to serialize v: %w", err) }).
		Serialize()
}
