// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// package adkg implements Async DKG as described in
//
// 		Sourav Das, Tom Yurek, Zhuolun Xiang, Andrew Miller, Lefteris Kokoris-Kogias,
// 		and Ling Ren. Practical Asynchronous Distributed Key Generation. Cryptology.
// 		https://eprint.iacr.org/2021/1591
//
// The algorithm adopted to the Feldman commitments is provided at
// <https://iotaledger.github.io/crypto-tss/talks/async-dkg/slides-async-dkg.html#/6/3>.
// The copy of the pseudocode follows:
//
// > // party i with input sᵢ
// > input sᵢ to ACSSᵢ
// >
// > on termination of the ACSSⱼ:
// >   sⱼ := output of ACSSⱼ
// >   Tᵢ = Tᵢ ∪ {j}
// >   if len(Tᵢ) == f + 1:
// >     RBCᵢ(Tᵢ)
// >
// > for j != i: RBCⱼ(Tⱼ) with predicate P(·)
// >
// > func P(Tⱼ):
// >   when Tⱼ ⊆ Tᵢ:
// >     return true
// >
// > on termination of RBCⱼ:
// >   when Tⱼ ⊆ Tᵢ:
// >     Input 1 to ABAⱼ
// >
// > on termination of ABAⱼ:
// >   if ABAⱼ outputs 1:
// >     T = T ∪ Tⱼ
// >     input 0 to all remaining ABAs
// >
// > wait until all ABAs terminate
// > z := sum(sⱼ for j in T)
// > output z
//
package adkg

import (
	"sort"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/gpa/acss"
	rbc "github.com/iotaledger/wasp/packages/gpa/rbc/bracha"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/suites"
	"golang.org/x/xerrors"
)

type adkgImpl struct {
	suite   suites.Suite
	n       int
	f       int
	me      gpa.NodeID
	myIdx   int
	nodeIDs []gpa.NodeID
	acss    []gpa.GPA
	rbcs    []gpa.GPA
	st      map[int]*share.PriShare // > Let Si = {}; Ti = {}
	log     *logger.Logger
}

var _ gpa.GPA = &adkgImpl{}

func New(
	suite suites.Suite,
	peers []gpa.NodeID,
	peerPKs map[gpa.NodeID]kyber.Point,
	f int,
	me gpa.NodeID,
	mySK kyber.Scalar,
	log *logger.Logger,
) gpa.GPA {
	n := len(peers)
	myIdx := -1
	for i := range peers {
		if peers[i] == me {
			myIdx = i
		}
	}
	if myIdx == -1 {
		panic("i'm not in the peer list")
	}
	a := &adkgImpl{
		suite:   suite,
		n:       n,
		f:       f,
		me:      me,
		myIdx:   myIdx,
		nodeIDs: peers,
		acss:    nil,                       // Will be set bellow.
		rbcs:    nil,                       // Will be set bellow.
		st:      map[int]*share.PriShare{}, // > Let Si = {}; Ti = {}
		log:     log,
	}
	a.acss = make([]gpa.GPA, len(peers))
	for i := range a.acss {
		a.acss[i] = acss.New(suite, peers, peerPKs, f, me, mySK, peers[i], nil, log)
	}
	a.rbcs = make([]gpa.GPA, len(peers))
	for i := range a.rbcs {
		a.rbcs[i] = rbc.New(peers, f, me, peers[i], a.rbcPredicate)
	}
	return gpa.NewOwnHandler(me, a)
}

func (a *adkgImpl) Input(input gpa.Input) []gpa.Message {
	if input != nil {
		panic(xerrors.Errorf("only expect a nil input, got: %+v", input))
	}
	secret := a.suite.Scalar().Pick(a.suite.RandomStream())
	return WrapMessages(msgWrapperACSS, a.myIdx, a.acss[a.myIdx].Input(secret))
}

func (a *adkgImpl) Message(msg gpa.Message) []gpa.Message {
	switch msgT := msg.(type) {
	case *msgWrapper:
		switch msgT.subsystem {
		case msgWrapperACSS:
			return a.handleACSSMessage(msgT)
		case msgWrapperRBC:
			return WrapMessages(msgWrapperRBC, msgT.index, a.rbcs[msgT.index].Message(msgT.wrapped))
		}
	case *msgACSSOutput:
		return a.handleACSSOutput(msgT)
	default:
		panic(xerrors.Errorf("unexpected message: %+v", msg))
	}
	return nil
}

func (a *adkgImpl) Output() gpa.Output {
	return nil
}

func (a *adkgImpl) handleACSSMessage(msg *msgWrapper) []gpa.Message {
	msgs := WrapMessages(msgWrapperACSS, msg.index, a.acss[msg.index].Message(msg.wrapped))
	out := a.acss[msg.index].Output()
	if out != nil && a.st[msg.index] == nil {
		priShare, ok := out.(*share.PriShare)
		if !ok {
			panic(xerrors.Errorf("acss output wrong type: %+v", out))
		}
		msgs = append(msgs, &msgACSSOutput{me: a.me, index: msg.index, priShare: priShare})
	}
	return msgs
}

//
// > on termination of the ACSSⱼ:
// >   sⱼ := output of ACSSⱼ
// >   Tᵢ = Tᵢ ∪ {j}
// >   if len(Tᵢ) == f + 1:
// >     RBCᵢ(Tᵢ)
//
func (a *adkgImpl) handleACSSOutput(acssOutput *msgACSSOutput) []gpa.Message {
	j := acssOutput.index
	if _, ok := a.st[j]; ok {
		// Already set. Ignore the duplicate messages.
		return gpa.NoMessages()
	}
	a.st[j] = acssOutput.priShare
	if len(a.st) == a.f+1 {
		t := make([]int, 0)
		for ti := range a.st {
			t = append(t, ti)
		}
		sort.Ints(t)
		payloadBytes, err := (&msgRBCPayload{t: t}).MarshalBinary()
		if err != nil {
			panic(xerrors.Errorf("cannot serialize RBC payload: %v", err))
		}
		for i := range a.rbcs {
			rbc.SendPredicateUpdate(a.rbcs[i], a.me, a.rbcPredicate)
		}
		return a.rbcs[a.myIdx].Input(payloadBytes)
	}
	return gpa.NoMessages()
}

//
// > func P(Tⱼ):
// >   when Tⱼ ⊆ Tᵢ:
// >     return true
//
func (a *adkgImpl) rbcPredicate(rbcPayload []byte) bool {
	parsed := &msgRBCPayload{}
	if err := parsed.UnmarshalBinary(rbcPayload); err != nil {
		a.log.Warnf("cannot parse the rbc payload: %v", err)
		return false
	}
	for _, ti := range parsed.t {
		if _, ok := a.st[ti]; !ok {
			return false
		}
	}
	return true
}
