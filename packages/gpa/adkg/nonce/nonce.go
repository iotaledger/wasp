// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// nonce package implements NonceDKG as described in <https://github.com/iotaledger/crypto-tss/>.
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
//
package nonce

import (
	"fmt"
	"sort"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/gpa/acss"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/suites"
	"golang.org/x/xerrors"
)

type Output struct {
	Indexes  []int           // Indexes used to construct the final key (exactly f+1 for the intermediate output).
	PubKey   kyber.Point     // The common/aggregated public key of the key set.
	PriShare *share.PriShare // Final key share (can be nil until consensus is completed in the case of aggrExt==true).
	Commits  []kyber.Point   // Commitments for the final key shares.
}

type nonceDKGImpl struct {
	suite     suites.Suite
	n         int
	f         int
	me        gpa.NodeID
	myIdx     int
	nodeIDs   []gpa.NodeID
	acss      []gpa.GPA
	st        map[int]*share.PriShare // > Let Si = {}; Ti = {}
	stCommits map[int][]kyber.Point   // Commits for Si.
	agreedT   []int                   // Output from the external consensus.
	output    gpa.Output              // Output of the ADKG, can be intermediate (PriShare=nil).
	wrapper   *gpa.MsgWrapper
	log       *logger.Logger
}

var _ gpa.GPA = &nonceDKGImpl{}

const (
	msgWrapperACSS byte = iota // subsystem code.
)

const (
	msgTypeWrapped byte = iota
)

func New(
	suite suites.Suite,
	nodeIDs []gpa.NodeID,
	peerPKs map[gpa.NodeID]kyber.Point,
	f int,
	me gpa.NodeID,
	mySK kyber.Scalar,
	log *logger.Logger,
) gpa.GPA {
	myIdx := -1
	for i := range nodeIDs {
		if nodeIDs[i] == me {
			myIdx = i
		}
	}
	if myIdx == -1 {
		panic("i'm not in the peer list")
	}
	n := &nonceDKGImpl{
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
		wrapper:   nil,
		log:       log,
	}
	n.wrapper = gpa.NewMsgWrapper(msgTypeWrapped, n.subsystemFunc)
	n.acss = make([]gpa.GPA, len(nodeIDs))
	for i := range n.acss {
		n.acss[i] = acss.New(suite, nodeIDs, peerPKs, f, me, mySK, nodeIDs[i], nil, log)
	}
	return gpa.NewOwnHandler(me, n)
}

func (n *nonceDKGImpl) Input(input gpa.Input) []gpa.Message {
	if input != nil {
		panic(xerrors.Errorf("only expect a nil input, got: %+v", input))
	}
	secret := n.suite.Scalar().Pick(n.suite.RandomStream())
	msgs := n.wrapper.WrapMessages(msgWrapperACSS, n.myIdx, n.acss[n.myIdx].Input(secret))
	return n.tryHandleACSSTermination(n.myIdx, msgs)
}

func (n *nonceDKGImpl) Message(msg gpa.Message) []gpa.Message {
	switch msgT := msg.(type) {
	case *gpa.WrappingMsg:
		switch msgT.Subsystem() {
		case msgWrapperACSS:
			return n.handleACSSMessage(msgT)
		default:
			panic(xerrors.Errorf("unexpected message: %+v", msg))
		}
	case *msgACSSOutput:
		return n.handleACSSOutput(msgT)
	case *msgAgreementResult:
		return n.handleAgreementResult(msgT)
	default:
		panic(xerrors.Errorf("unexpected message: %+v", msg))
	}
}

func (n *nonceDKGImpl) Output() gpa.Output {
	return n.output
}

func (n *nonceDKGImpl) StatusString() string {
	acssStats := ""
	for i := range n.acss {
		acssStats += "\n" + n.acss[i].StatusString()
	}
	return fmt.Sprintf("{ADKG:Nonce, acss: %s}", acssStats)
}

func (n *nonceDKGImpl) handleACSSMessage(msg *gpa.WrappingMsg) []gpa.Message {
	msgIndex := msg.Index()
	msgs := n.wrapper.WrapMessages(msgWrapperACSS, msgIndex, n.acss[msgIndex].Message(msg.Wrapped()))
	return n.tryHandleACSSTermination(msgIndex, msgs)
}

func (n *nonceDKGImpl) tryHandleACSSTermination(acssIndex int, msgs []gpa.Message) []gpa.Message {
	out := n.acss[acssIndex].Output()
	if out != nil && n.st[acssIndex] == nil {
		acssOutput, ok := out.(*acss.Output)
		if !ok {
			panic(xerrors.Errorf("acss output wrong type: %+v", out))
		}
		msgs = append(msgs, &msgACSSOutput{me: n.me, index: acssIndex, priShare: acssOutput.PriShare, commits: acssOutput.Commits})
	}
	return msgs
}

func (n *nonceDKGImpl) handleACSSOutput(acssOutput *msgACSSOutput) []gpa.Message {
	j := acssOutput.index
	if _, ok := n.st[j]; ok {
		// Already set. Ignore the duplicate messages.
		return gpa.NoMessages()
	}
	n.st[j] = acssOutput.priShare
	n.stCommits[j] = acssOutput.commits
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

func (n *nonceDKGImpl) handleAgreementResult(msg *msgAgreementResult) []gpa.Message {
	if n.agreedT != nil {
		return gpa.NoMessages()
	}

	if len(msg.proposals) < n.n-n.f {
		panic(xerrors.Errorf("len(msg.proposals) < n.n - n.f, len=%v, n=%v, f=%v", len(msg.proposals), n.n, n.f))
	}
	voteCounts := make([]int, n.n)
	for _, proposal := range msg.proposals {
		if len(proposal) < n.f+1 {
			n.log.Warnf("len(proposal) < f+1, that should not happen")
			continue
		}
		for i := range proposal {
			duplicatesFound := false
			for j := range proposal {
				if i != j && proposal[i] == proposal[j] {
					duplicatesFound = true
					n.log.Warnf("msgAgreementResult with duplicate votes")
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
		panic(xerrors.Errorf("len(agreedT) < f+1, that should not happen, len=%v, f=%v", len(agreedT), n.f))
	}
	n.agreedT = agreedT
	return n.tryMakeFinalOutput()
}

func (n *nonceDKGImpl) tryMakeFinalOutput() []gpa.Message {
	if n.agreedT == nil {
		return gpa.NoMessages()
	}
	var sumCommitPoly *share.PubPoly
	sum := n.suite.Scalar().Zero()
	for _, j := range n.agreedT {
		if _, ok := n.st[j]; !ok {
			return gpa.NoMessages()
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
				n.log.Error("Unable to sum public commitments: %v", err)
				return gpa.NoMessages()
			}
		}
	}
	_, sumCommit := sumCommitPoly.Info()
	n.output = &Output{
		Indexes:  n.agreedT,
		PubKey:   sumCommit[0],
		PriShare: &share.PriShare{I: n.myIdx, V: sum},
		Commits:  sumCommit,
	}
	return gpa.NoMessages()
}
