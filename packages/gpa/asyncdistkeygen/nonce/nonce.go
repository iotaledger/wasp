// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package nonce implements NonceDKG as described in <https://github.com/iotaledger/crypto-tss/>.
// > 4) Asynchronous nonce-DKG
// > Variant a)
// >
// >     Setup
// >         Run any DKG (preferably probably FROST-DKG) to derive the aggregated public key and private key share.
// >         This leads to a synchronous, non-robust setup phase.
// >     Nonce sharing (can be started any time before the signing process)
// >         For every party i:
// >             Sample secret s = a₀
// >             Run ACSSᵢ(s):
// >                 C=(A₀,A₁,…,Aₜ), e=(Enc_pk₀(y₀),…,Enc_pkₙ(yₙ)) ← VSSEncAndProve(s)
// >                 Broadcast (C,e) using Verified Reliable Broadcast (RBC) with predicate: C is valid
// >             On termination of ACSSⱼ:
// >                 sʲᵢ ← output
// >                 Tᵢ ← Tᵢ ∪ {j}
// >             Wait until |Tᵢ| ≥ n - f
// >     Signing process
// >         For every party i:
// >             Input Tᵢ (bit vector) into Verified ACS with predicate: |Tᵢ| ≥ n - f
// >             On termination of ACS:
// >                 𝒯 ← {j | the j-th bit is set in at least f+1 elements of the output}
// >                 (One can show that |𝒯| ≥ f + 1 will always hold. Thus, one honest dealer will always be included.)
// >                 Wait until 𝒯 ⊆ Tᵢ
// >                 (as for each j in 𝒯 at least one honest peer observed a termination of ACSSⱼ, this will eventually succeed.)
// >                 σᵢ ← sum(sʲᵢ for j in 𝒯)
// >             Create partial signature using the private key share and σᵢ as the nonce share
// >         Aggregate t partial signatures to form the valid signature
package nonce

import (
	"fmt"
	"slices"
	"sort"

	"github.com/samber/lo"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/suites"

	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/gpa/acss"
	rbc "github.com/iotaledger/wasp/v2/packages/gpa/rbc/bracha"
)

type Output struct {
	Indexes   []int           // Indexes used to construct the final key (exactly f+1 for the intermediate output).
	PubKey    kyber.Point     // The common/aggregated public key of the key set.
	PriShare  *share.PriShare // Final key share (can be nil until consensus is completed in the case of aggrExt==true).
	Commits   []kyber.Point   // Commitments for the final key shares.
	Threshold int
}

type NonceDistributedKeyGeneration struct {
	suite     suites.Suite
	n         int
	f         int
	me        gpa.NodeID
	myIdx     int
	nodeIDs   []gpa.NodeID
	acss      []*acss.ACSS
	st        map[int]*share.PriShare // > Let Si = {}; Ti = {}
	stCommits map[int][]kyber.Point   // Commits for Si.
	agreedT   []int                   // Output from the external consensus.
	output    gpa.Output              // Output of the ADKG, can be intermediate (PriShare=nil).
	log       log.Logger
}

const (
	SubsystemID string = "nonce"
)

func New(
	suite suites.Suite,
	nodeIDs []gpa.NodeID,
	peerPKs map[gpa.NodeID]kyber.Point,
	f int,
	me gpa.NodeID,
	mySK kyber.Scalar,
	log log.Logger,
) *NonceDistributedKeyGeneration {
	myIdx := -1
	for i := range nodeIDs {
		if nodeIDs[i] == me {
			myIdx = i
		}
	}
	if myIdx == -1 {
		panic("i'm not in the peer list")
	}
	n := &NonceDistributedKeyGeneration{
		suite:     suite,
		n:         len(nodeIDs),
		f:         f,
		me:        me,
		myIdx:     myIdx,
		nodeIDs:   nodeIDs,
		acss:      nil,                       // Will be set bellow.
		st:        map[int]*share.PriShare{}, // > Let Si = {}; Ti = {}
		stCommits: map[int][]kyber.Point{},   // Commits for Si/Ti.
		agreedT:   nil,                       // Will be set, when output from the consensus will be received.
		output:    nil,                       // Can be intermediate (PriShare == nil) or final (PriShare != nil).
		log:       log,
	}
	n.acss = make([]*acss.ACSS, len(nodeIDs))
	for i := range n.acss {
		n.acss[i] = acss.New(suite, nodeIDs, peerPKs, f, me, mySK, nodeIDs[i], nil, log)
	}
	return n
}

func (n *NonceDistributedKeyGeneration) Input(input gpa.Input) []gpa.MessageOut {
	switch input := input.(type) {
	case *inputStart:
		secret := n.suite.Scalar().Pick(n.suite.RandomStream())
		msgs := gpa.AddKey(SubsystemID, n.myIdx, n.acss[n.myIdx].Input(secret))
		return slices.Concat(msgs, n.tryHandleACSSTermination(n.myIdx))
	case *inputAgreementResult:
		return n.handleAgreementResult(input)
	}
	panic(fmt.Errorf("unexpected input %T: %+v", input, input))
}

// func (n *NonceDistributedKeyGeneration) wrapACSSMessagesOut(acssIndex int, msgs acss.[]gpa.PayloadOut) []gpa.PayloadOut {
// 	wrappedMsgs := &[]gpa.PayloadOutV{}
// 	wrappedMsgs.Vote = lo.Map(msgs.Vote, func(m gpa.PayloadOut[acss.MsgVote], _ int) gpa.PayloadOut[MsgACSSVote] {
// 		return gpa.NewPayloadOut[MsgACSSVote](m.Recipient, MsgACSSVote{Index: acssIndex, MsgVote: m.Payload})
// 	})
// 	wrappedMsgs.RBCCEPayload = lo.Map(msgs.RBCCEPayload, func(m gpa.PayloadOut[acss.MsgRBCCEPayload], _ int) gpa.PayloadOut[MsgACSSRBCCEPayload] {
// 		return gpa.NewPayloadOut[MsgACSSRBCCEPayload](m.Recipient, MsgACSSRBCCEPayload{Index: acssIndex, MsgRBCCEPayload: m.Payload})
// 	})
// 	wrappedMsgs.ImplicateRecover = lo.Map(msgs.ImplicateRecover, func(m gpa.PayloadOut[acss.MsgImplicateRecover], _ int) gpa.PayloadOut[MsgACSSImplicateRecover] {
// 		return gpa.NewPayloadOut[MsgACSSImplicateRecover](m.Recipient, MsgACSSImplicateRecover{Index: acssIndex, MsgImplicateRecover: m.Payload})
// 	})
// 	wrappedMsgs.Bracha = lo.Map(msgs.Bracha, func(m gpa.PayloadOut[rbc.MsgBracha], _ int) gpa.PayloadOut[MsgACSSBracha] {
// 		return gpa.NewPayloadOut[MsgACSSBracha](m.Recipient, MsgACSSBracha{Index: acssIndex, MsgBracha: m.Payload})
// 	})
// 	return wrappedMsgs
// }

func (n *NonceDistributedKeyGeneration) Output() gpa.Output {
	return n.output
}

func (n *NonceDistributedKeyGeneration) StatusString() string {
	acssStats := ""
	for i := range n.acss {
		acssStats += "\n" + n.acss[i].StatusString()
	}
	return fmt.Sprintf("{ADKG:Nonce, acss: %s}", acssStats)
}

func (n *NonceDistributedKeyGeneration) HandleACSSMsgBracha(acssIndex int, m gpa.MessageIn[rbc.MsgBracha]) []gpa.MessageOut {
	outMsgs := n.acss[acssIndex].HandleRBCMsgBracha(m)
	return slices.Concat(
		gpa.AddKey(SubsystemID, acssIndex, outMsgs),
		n.tryHandleACSSTermination(acssIndex),
	)
}

func (n *NonceDistributedKeyGeneration) HandleACSSMsgVote(acssIndex int, msg gpa.MessageIn[acss.MsgVote]) []gpa.MessageOut {
	outMsgs := n.acss[acssIndex].HandleMsgVote(msg)
	return slices.Concat(
		gpa.AddKey(SubsystemID, acssIndex, outMsgs),
		n.tryHandleACSSTermination(acssIndex),
	)
}

func (n *NonceDistributedKeyGeneration) HandleACSSMsgImplicateRecover(acssIndex int, msg gpa.MessageIn[acss.MsgImplicateRecover]) []gpa.MessageOut {
	outMsgs := n.acss[acssIndex].HandleImplicateRecoverReceived(msg)
	return slices.Concat(
		gpa.AddKey(SubsystemID, acssIndex, outMsgs),
		n.tryHandleACSSTermination(acssIndex),
	)
}

func (n *NonceDistributedKeyGeneration) tryHandleACSSTermination(acssIndex int) []gpa.MessageOut {
	out := n.acss[acssIndex].Output()
	if out != nil && n.st[acssIndex] == nil {
		acssOutput, ok := out.(*acss.Output)
		if !ok {
			panic(fmt.Errorf("acss output wrong type: %+v", out))
		}
		return n.handleACSSOutput(acssIndex, acssOutput.PriShare, acssOutput.Commits)
	}
	return nil
}

func (n *NonceDistributedKeyGeneration) handleACSSOutput(index int, priShare *share.PriShare, commits []kyber.Point) []gpa.MessageOut {
	j := index
	if _, ok := n.st[j]; ok {
		// Already set. Ignore the duplicate messages.
		return nil
	}
	n.st[j] = priShare
	n.stCommits[j] = commits
	if len(n.st) == n.n-n.f && n.output == nil {
		t := make([]int, 0)
		for ti := range n.st {
			t = append(t, ti)
		}
		sort.Ints(t)
		n.output = &Output{Indexes: t} // That's intermediate output.
	}
	//
	// It is possible that the indexes are already decided and are waiting for the ACSS only.
	// Thus we have to try produce the final output.
	return n.tryMakeFinalOutput()
}

func (n *NonceDistributedKeyGeneration) handleAgreementResult(input *inputAgreementResult) []gpa.MessageOut {
	if n.agreedT != nil {
		return nil
	}

	if len(input.proposals) < n.n-n.f {
		panic(fmt.Errorf("len(msg.proposals) < n.n - n.f, len=%v, n=%v, f=%v", len(input.proposals), n.n, n.f))
	}
	voteCounts := make([]int, n.n)
	for _, proposal := range input.proposals {
		if len(proposal) < n.f+1 {
			n.log.LogWarn("len(proposal) < f+1, that should not happen")
			continue
		}
		for i := range proposal {
			duplicatesFound := false
			for j := range proposal {
				if i != j && proposal[i] == proposal[j] {
					duplicatesFound = true
					n.log.LogWarn("msgAgreementResult with duplicate votes")
				}
			}
			if !duplicatesFound {
				voteCounts[proposal[i]]++
			}
		}
	}
	agreedT := []int{}
	for i := range voteCounts {
		if voteCounts[i] >= n.f+1 {
			agreedT = append(agreedT, i)
		}
	}
	if len(agreedT) < n.f+1 {
		panic(fmt.Errorf("len(agreedT) < f+1, that should not happen, len=%v, f=%v", len(agreedT), n.f))
	}
	n.agreedT = agreedT
	return n.tryMakeFinalOutput()
}

func (n *NonceDistributedKeyGeneration) tryMakeFinalOutput() []gpa.MessageOut {
	if n.agreedT == nil {
		return nil
	}
	var sumCommitPoly *share.PubPoly
	sum := n.suite.Scalar().Zero()
	for _, j := range n.agreedT {
		if _, ok := n.st[j]; !ok {
			n.log.LogDebugf("Don't have S/T[%v] yet, have to wait, agreedT=%+v, have S/T indexes: %v.", j, n.agreedT, lo.Keys(n.st))
			return nil
		}
		sum.Add(sum.Clone(), n.st[j].V)
		//
		// Sum the polynomials as well.
		jCommitPoly := share.NewPubPoly(n.suite, nil, n.stCommits[j])
		if sumCommitPoly == nil {
			sumCommitPoly = jCommitPoly
		} else {
			var err error
			sumCommitPoly, err = sumCommitPoly.Add(jCommitPoly)
			if err != nil {
				n.log.LogError("Unable to sum public commitments: %v", err)
				return nil
			}
		}
	}
	_, sumCommit := sumCommitPoly.Info()
	n.output = &Output{
		Indexes:   n.agreedT,
		PubKey:    sumCommit[0],
		PriShare:  &share.PriShare{I: n.myIdx, V: sum},
		Commits:   sumCommit,
		Threshold: n.n - n.f,
	}
	return nil
}
