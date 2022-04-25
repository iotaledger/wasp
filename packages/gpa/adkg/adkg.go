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
// >     Input 1 to ABAⱼ					\
// >										|
// > on termination of ABAⱼ:				| NOTE: This is an agreement on the set of
// >   if ABAⱼ outputs 1:					| indexes to include into the final share.
// >     T = T ∪ Tⱼ							| That is made optional by allowing a user
// >     input 0 to all remaining ABAs		| to provide a decision (see aggrExt parameter).
// >										|
// > wait until all ABAs terminate			/
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

type Output struct {
	Indexes  []int           // Intexes used to construct the final key (exactly f+1 for the intermediate output).
	PriShare *share.PriShare // Final key share (can be nil until consensus is completed in the case of aggrExt==true).
}

type adkgImpl struct {
	suite   suites.Suite
	n       int
	f       int
	me      gpa.NodeID
	myIdx   int
	nodeIDs []gpa.NodeID
	aggrExt bool // Should we delegate the agreement on share set to the user?
	acss    []gpa.GPA
	rbcs    []gpa.GPA
	st      map[int]*share.PriShare // > Let Si = {}; Ti = {}
	agreedT []int                   // > T = T ∪ Tⱼ
	output  *Output                 // Output of the ADKG, can be intermediate, if aggrExt==true.
	log     *logger.Logger
}

var _ gpa.GPA = &adkgImpl{}

func New(
	suite suites.Suite,
	nodeIDs []gpa.NodeID,
	peerPKs map[gpa.NodeID]kyber.Point,
	aggrExt bool, // Should we use external agreement?
	f int,
	me gpa.NodeID,
	mySK kyber.Scalar,
	log *logger.Logger,
) gpa.GPA {
	if !aggrExt {
		panic(xerrors.Errorf("internal agreement not implemented yet")) // TODO: Implement it.
	}
	n := len(nodeIDs)
	myIdx := -1
	for i := range nodeIDs {
		if nodeIDs[i] == me {
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
		nodeIDs: nodeIDs,
		aggrExt: aggrExt,
		acss:    nil,                       // Will be set bellow.
		rbcs:    nil,                       // Will be set bellow.
		st:      map[int]*share.PriShare{}, // > Let Si = {}; Ti = {}
		output:  nil,
		log:     log,
	}
	a.acss = make([]gpa.GPA, len(nodeIDs))
	for i := range a.acss {
		a.acss[i] = acss.New(suite, nodeIDs, peerPKs, f, me, mySK, nodeIDs[i], nil, log)
	}
	a.rbcs = make([]gpa.GPA, len(nodeIDs))
	for i := range a.rbcs {
		a.rbcs[i] = rbc.New(nodeIDs, f, me, nodeIDs[i], a.rbcPredicate)
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
			return a.handleRBCMessage(msgT)
		default:
			panic(xerrors.Errorf("unexpected message: %+v", msg))
		}
	case *msgACSSOutput:
		return a.handleACSSOutput(msgT)
	case *msgAgreementResult:
		return a.handleAgreementResult(msgT)
	default:
		panic(xerrors.Errorf("unexpected message: %+v", msg))
	}
}

func (a *adkgImpl) Output() gpa.Output {
	return a.output
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
	msgs := []gpa.Message{}
	if len(a.st) >= a.f+1 {
		msgs = append(msgs, a.tryInput1ToABA()...)
		msgs = append(msgs, a.tryMakeFinalOutput()...)
		//
		// Update the predicates.
		for i := range a.rbcs {
			msgs = append(msgs, WrapMessage(msgWrapperRBC, i, rbc.MakePredicateUpdateMsg(a.me, a.rbcPredicate)))
		}
	}
	if len(a.st) == a.f+1 {
		//
		// Provide input to our RBC.
		t := make([]int, 0)
		for ti := range a.st {
			t = append(t, ti)
		}
		sort.Ints(t)
		payloadBytes, err := (&msgRBCPayload{t: t}).MarshalBinary()
		if err != nil {
			panic(xerrors.Errorf("cannot serialize RBC payload: %v", err))
		}
		msgs = append(msgs, WrapMessages(msgWrapperRBC, a.myIdx, a.rbcs[a.myIdx].Input(payloadBytes))...)
	}
	return msgs
}

func (a *adkgImpl) handleRBCMessage(msg *msgWrapper) []gpa.Message {
	msgs := WrapMessages(msgWrapperRBC, msg.index, a.rbcs[msg.index].Message(msg.wrapped))
	out := a.rbcs[msg.index].Output()
	if out != nil {
		// > on termination of RBCⱼ:
		// >   when Tⱼ ⊆ Tᵢ:
		// >     Input 1 to ABAⱼ
		return append(msgs, a.tryInput1ToABA()...)
	}
	return msgs
}

// This event can happen because of the termination of all ABAs (if aggrExt==false)
// or can be sent by the user (if aggrExt== true).
//
// > wait until all ABAs terminate
// > z := sum(sⱼ for j in T)
// > output z
//
// NOTE: Here we also have to wait for ACSS instances to terminate, whose indexes
// are in the decided set of shares.
//
func (a *adkgImpl) handleAgreementResult(msg *msgAgreementResult) []gpa.Message {
	if a.agreedT != nil {
		return gpa.NoMessages()
	}
	a.agreedT = msg.indexes
	return a.tryMakeFinalOutput()
}

// The following condition can become true upon reception of RBCⱼ output or
// when Tᵢ increases. This function is called in both cases.
//
// > on termination of RBCⱼ:
// >   when Tⱼ ⊆ Tᵢ:
// >     Input 1 to ABAⱼ
//
func (a *adkgImpl) tryInput1ToABA() []gpa.Message { // TODO: Make the check more efficient. Will be called a lot.
	for _, r := range a.rbcs {
		ro := r.Output()
		if ro == nil {
			continue
		}
		roBytes, ok := ro.([]byte)
		if !ok {
			panic(xerrors.Errorf("unexpected RBC output: %+v", ro))
		}
		if a.rbcPredicate(roBytes) {
			rbcPayload := &msgRBCPayload{}
			if err := rbcPayload.UnmarshalBinary(roBytes); err != nil {
				panic(xerrors.Errorf("cannot decode rbc payload: %v", err))
			}
			if a.aggrExt {
				if a.output == nil {
					//
					// In this case we are returning an intermediate output for user
					// to on the index set to use for the final secret share. I.e.
					// An external consensus is assumed in this case.
					//
					a.output = &Output{
						Indexes:  rbcPayload.t, // Take first RBC output as a chosen candidate.
						PriShare: nil,          // That's intermediate result.
					}
				}
				continue
			}
			panic(xerrors.Errorf("internal ABA not implemented yet")) // TODO: Input 1 to ABA.
		}
	}
	return gpa.NoMessages()
}

// Final decision can be made when all the agreement is terminated (all ABAs terminated
// or user supplied decision provided) or when the missing ACSS is terminated.
//
// > wait until all ABAs terminate
// > z := sum(sⱼ for j in T)
// > output z
//
// NOTE: Here we also have to wait for ACSS instances to terminate, whose indexes
// are in the decided set of shares.
func (a *adkgImpl) tryMakeFinalOutput() []gpa.Message {
	if a.agreedT == nil {
		return gpa.NoMessages()
	}
	sum := a.suite.Scalar().Zero()
	for _, j := range a.agreedT {
		if _, ok := a.st[j]; !ok {
			return gpa.NoMessages()
		}
		sum.Add(sum.Clone(), a.st[j].V)
	}
	a.output = &Output{
		Indexes:  a.agreedT,
		PriShare: &share.PriShare{I: a.myIdx, V: sum},
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
