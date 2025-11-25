// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Package acss implements "Asynchronous Complete Secret Sharing" as described in
//
//	https://iotaledger.github.io/crypto-tss/talks/async-dkg/slides-async-dkg.html#/5/6
//
// Here is a copy of the pseudo code from the slide mentioned above (just in case):
//
// > // dealer with input s
// > sample random polynomial ϕ such that ϕ(0) = s
// > C, S := VSS.Share(ϕ, f+1, n)
// > E := [PKI.Enc(S[i], pkᵢ) for each party i]
// >
// > // party i (including the dealer)
// > RBC(C||E)
// > sᵢ := PKI.Dec(eᵢ, skᵢ)
// > if decrypt fails or VSS.Verify(C, i, sᵢ) == false:
// >   send <IMPLICATE, i, skᵢ> to all parties
// > else:
// >   send <OK>
// >
// > on receiving <OK> from n-f parties:
// >   send <READY> to all parties
// >
// > on receiving <READY> from f+1 parties:
// >   send <READY> to all parties
// >
// > on receiving <READY> from n-f parties:
// >   if sᵢ is valid:
// >     out = true
// >     output sᵢ
// >
// > on receiving <IMPLICATE, j, skⱼ>:
// >   sⱼ := PKI.Dec(eⱼ, skⱼ)
// >   if decrypt fails or VSS.Verify(C, j, sⱼ) == false:
// >     if out == true:
// >       send <RECOVER, i, skᵢ> to all parties
// >       return
// >
// >     on receiving <RECOVER, j, skⱼ>:
// >       sⱼ := PKI.Dec(eⱼ, skⱼ)
// >       if VSS.Verify(C, j, sⱼ): T = T ∪ {sⱼ}
// >
// >     wait until len(T) >= f+1:
// >       sᵢ = SSS.Recover(T, f+1, n)(i)
// >       out = true
// >       output sᵢ
//
// On the adaptations and sources:
//
// > More details and references to the papers are bellow:
// >
// > Here the references for the Asynchronous Secret-Sharing that I was referring to.
// > It is purely based on (Feldman) Verifiable Secret Sharing and does not rely on any PVSS schemes
// > requiring fancy NIZKP (and thus trades network-complexity vs computational-complexity):
// >
// >   * [1], Section IV. A. we use the ACSS scheme from [2] but replace its Pedersen
// >     commitment with a Feldman polynomial commitment to achieve Homomorphic-Partial-Commitment.
// >
// >   * In [2], Section 5.3. they explain the Pedersen-based hbACSS0 and give some proof sketch.
// >     The complete description and analysis of hbACSS0 can be found in [3]. However, as mentioned
// >     before they use Kate-commitments instead of Feldman/Pedersen. This has better message
// >     complexity especially when multiple secrets are shared at the same time, but in our case
// >     that would need to be replaced with Feldman making it much simpler and not losing any security.
// >     Actually, [3] is just a pre-print, the official published version is [4], but [4] also contains
// >     other, non-relevant, variants like hbACSS1 and hbACSS2 and much more analysis.
// >     So, I found [3] a bit more helpful, although it is just the preliminary version.
// >     They also provide their reference implementation in [5], which is also what the
// >     authors of [1] used for their practical DKG results.
// >
// > [1] Practical Asynchronous Distributed Key Generation https://eprint.iacr.org/2021/1591
// > [2] Asynchronous Data Dissemination and its Applications https://eprint.iacr.org/2021/777
// > [3] Brief Note: Asynchronous Verifiable Secret Sharing with Optimal Resilience and Linear Amortized Overhead https://arxiv.org/pdf/1902.06095.pdf
// > [4] hbACSS: How to Robustly Share Many Secrets https://eprint.iacr.org/2021/159
// > [5] https://github.com/tyurek/hbACSS
//
// A PoC implementation: <https://github.com/Wollac/async.go>.
//
// The Crypto part shown the pseudo-code above is replaced in the implementation with the
// scheme allowing to keep the private keys secret. The scheme implementation is taken
// from the PoC mentioned above. It is described in <https://hackmd.io/@CcRtfCBnRbW82-AdbFJUig/S1qcPiUN5>.
package acss

import (
	"errors"
	"fmt"
	"math"
	"slices"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/suites"

	"github.com/samber/lo"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/gpa/acss/crypto"
	"github.com/iotaledger/wasp/v2/packages/gpa/rbc/bracha"
	rbc "github.com/iotaledger/wasp/v2/packages/gpa/rbc/bracha"
)

const (
	subsystemRBC byte = iota
)

type Output struct {
	PriShare *share.PriShare // Private share, received by this instance.
	Commits  []kyber.Point   // Feldman's commitment to the shared polynomial.
}

type ACSS struct {
	suite         suites.Suite
	n             int
	f             int
	me            gpa.NodeID
	mySK          kyber.Scalar
	myPK          kyber.Point
	myIdx         int
	dealer        gpa.NodeID                           // A node that is recognized as a dealer.
	dealCB        func(int, []byte) []byte             // Callback to be called on the encrypted deals (for tests actually).
	peerPKs       map[gpa.NodeID]kyber.Point           // Peer public keys.
	peerIdx       []gpa.NodeID                         // Particular order of the nodes (position in the polynomial).
	rbc           *bracha.RBC                          // RBC to share `C||E`.
	rbcOut        *crypto.Deal                         // Deal broadcasted by the dealer.
	voteOKRecv    map[gpa.NodeID]bool                  // A set of received OK votes.
	voteREADYRecv map[gpa.NodeID]bool                  // A set of received READY votes.
	voteREADYSent bool                                 // Have we sent our READY vote?
	pendingIRMsgs []gpa.MessageIn[MsgImplicateRecover] // I/R messages are buffered, if the RBC is not completed yet.
	implicateRecv map[gpa.NodeID]bool                  // To check, that implicate only received once from a node.
	recoverRecv   map[gpa.NodeID]*share.PriShare       // Private shares from the RECOVER messages.
	outS          *share.PriShare                      // Our share of the secret (decrypted from rbcOutE).
	output        bool
	log           log.Logger
}

func New(
	suite suites.Suite, // Ed25519
	peers []gpa.NodeID, // Participating nodes in a specific order.
	peerPKs map[gpa.NodeID]kyber.Point, // Public keys for all the peers.
	f int, // Max number of expected faulty nodes.
	me gpa.NodeID, // ID of this node.
	mySK kyber.Scalar, // Secret Key of this node.
	dealer gpa.NodeID, // The dealer node for this protocol instance.
	dealCB func(int, []byte) []byte, // For tests only: interceptor for the deal to be shared.
	log log.Logger, // A logger to use.
) *ACSS {
	n := len(peers)
	if dealCB == nil {
		dealCB = func(i int, b []byte) []byte { return b }
	}
	a := ACSS{
		suite:         suite,
		n:             n,
		f:             f,
		me:            me,
		mySK:          mySK,
		myPK:          peerPKs[me],
		myIdx:         -1, // Updated bellow.
		dealer:        dealer,
		dealCB:        dealCB,
		peerPKs:       peerPKs,
		peerIdx:       peers,
		rbc:           rbc.New(peers, f, me, dealer, math.MaxInt, func(b []byte) bool { return true }, log), // TODO: Provide meaningful maxMsgSize
		rbcOut:        nil,                                                                                  // Will be set on output from the RBC.
		voteOKRecv:    map[gpa.NodeID]bool{},
		voteREADYRecv: map[gpa.NodeID]bool{},
		voteREADYSent: false,
		pendingIRMsgs: []gpa.MessageIn[MsgImplicateRecover]{},
		implicateRecv: map[gpa.NodeID]bool{},
		recoverRecv:   map[gpa.NodeID]*share.PriShare{},
		outS:          nil,
		output:        false,
		log:           log,
	}
	if a.myIdx = a.peerIndex(me); a.myIdx == -1 {
		panic("i'm not in the peer list")
	}
	return &a
}

// Input for the algorithm is the secret to share.
// It can be provided by the dealer only.
func (a *ACSS) Input(input gpa.Input) []gpa.MessageOut {
	if a.me != a.dealer {
		panic(errors.New("only dealer can initiate the sharing"))
	}
	if input == nil {
		panic(errors.New("we expect kyber.Scalar as input"))
	}
	return a.handleInput(input.(kyber.Scalar))
}

func (a *ACSS) HandleMsgVote(msg gpa.MessageIn[MsgVote]) []gpa.MessageOut {
	switch msg.Payload.kind {
	case msgVoteOK:
		return a.handleVoteOK(msg)
	case msgVoteREADY:
		return a.handleVoteREADY(msg)
	default:
		a.log.LogWarnf("unexpected vote message: %+v", msg)
		return nil
	}
}

// > // dealer with input s
// > sample random polynomial ϕ such that ϕ(0) = s
// > C, S := VSS.Share(ϕ, f+1, n)
// > E := [PKI.Enc(S[i], pkᵢ) for each party i]
// >
// > // party i (including the dealer)
// > RBC(C||E)
func (a *ACSS) handleInput(secretToShare kyber.Scalar) []gpa.MessageOut {
	pubKeys := make([]kyber.Point, 0)
	for _, peerID := range a.peerIdx {
		pubKeys = append(pubKeys, a.peerPKs[peerID])
	}
	deal := crypto.NewDeal(a.suite, pubKeys, secretToShare)
	data, err := deal.MarshalBinary()
	if err != nil {
		panic(fmt.Sprintf("acss: internal error: %v", err))
	}

	// > RBC(C||E)
	rbcCEPayloadBytes := bcs.MustMarshal(&MsgRBCCEPayload{data: data})
	msgs := a.rbc.Input(rbcCEPayloadBytes)
	return slices.Concat(msgs, a.tryHandleRBCTermination(false))
}

// Delegate received messages to the RBC and handle its output.
//
// > // party i (including the dealer)
// > RBC(C||E)
func (a *ACSS) HandleRBCMsgBracha(m gpa.MessageIn[rbc.MsgBracha]) []gpa.MessageOut {
	wasOut := a.rbc.Output() != nil // To send the msgRBCCEOutput message once (for perf reasons).
	msgs := a.rbc.HandleMsgBracha(m)
	return slices.Concat(msgs, a.tryHandleRBCTermination(wasOut))
}

func (a *ACSS) tryHandleRBCTermination(wasOut bool) []gpa.MessageOut {
	if out := a.rbc.Output(); !wasOut && out != nil {
		// Send the result for self as a message (maybe the code will look nicer this way).
		outParsed, err := bcs.UnmarshalInto(out.([]byte), &MsgRBCCEPayload{})
		if err != nil {
			outParsed = &MsgRBCCEPayload{}
		}
		return a.handleRBCOutput(outParsed, err)
	}
	return nil
}

// Upon receiving the RBC output...
//
// > sᵢ := PKI.Dec(eᵢ, skᵢ)
// > if decrypt fails or VSS.Verify(C, i, sᵢ) == false:
// >   send <IMPLICATE, i, skᵢ> to all parties
// > else:
// >   send <OK>
func (a *ACSS) handleRBCOutput(rbcOutput *MsgRBCCEPayload, err error) []gpa.MessageOut {
	if a.outS != nil || a.rbcOut != nil {
		// Take the first RBC output only.
		return nil
	}
	//
	// Store the broadcast result and process pending IMPLICATE/RECOVER messages, if any.
	if err != nil {
		return a.broadcastImplicate(err)
	}
	deal, err := crypto.DealUnmarshalBinary(a.suite, a.n, rbcOutput.data)
	if err != nil {
		return a.broadcastImplicate(errors.New("cannot unmarshal msgRBCCEPayload.data"))
	}
	a.rbcOut = deal
	msgs := a.handleImplicateRecoverPending()
	//
	// Process the RBC output, as described above.
	secret := crypto.Secret(a.suite, a.rbcOut.PubKey, a.mySK)
	myShare, err := crypto.DecryptShare(a.suite, a.rbcOut, a.myIdx, secret)
	if err != nil {
		return slices.Concat(msgs, a.broadcastImplicate(err))
	}
	a.outS = myShare
	a.tryOutput() // Maybe the READY messages are already received.
	return slices.Concat(
		msgs,
		a.broadcastVote(msgVoteOK),
		a.handleImplicateRecoverPending(),
	)
}

// > on receiving <OK> from n-f parties:
// >   send <READY> to all parties
func (a *ACSS) handleVoteOK(msg gpa.MessageIn[MsgVote]) []gpa.MessageOut {
	a.voteOKRecv[msg.Sender] = true
	count := len(a.voteOKRecv)
	if !a.voteREADYSent && count >= (a.n-a.f) {
		a.voteREADYSent = true
		return a.broadcastVote(msgVoteREADY)
	}
	return nil
}

// > on receiving <READY> from f+1 parties:
// >   send <READY> to all parties
// >
// > on receiving <READY> from n-f parties:
// >   if sᵢ is valid:
// >     out = true
// >     output sᵢ
func (a *ACSS) handleVoteREADY(msg gpa.MessageIn[MsgVote]) []gpa.MessageOut {
	a.voteREADYRecv[msg.Sender] = true
	count := len(a.voteREADYRecv)
	var msgs []gpa.MessageOut
	if !a.voteREADYSent && count >= (a.f+1) {
		msgs = a.broadcastVote(msgVoteREADY)
		a.voteREADYSent = true
	}
	a.tryOutput()
	return slices.Concat(msgs, a.handleImplicateRecoverPending())
}

// It is possible that we are receiving IMPLICATE/RECOVER messages before our RBC is completed.
// We store these messages for processing after that, if RBC is not done and process it otherwise.
func (a *ACSS) HandleImplicateRecoverReceived(msg gpa.MessageIn[MsgImplicateRecover]) []gpa.MessageOut {
	if a.rbcOut == nil {
		a.pendingIRMsgs = append(a.pendingIRMsgs, msg)
		return nil
	}
	switch msg.Payload.kind {
	case msgImplicateRecoverKindIMPLICATE:
		return a.handleImplicate(msg)
	case msgImplicateRecoverKindRECOVER:
		return a.handleRecover(msg)
	default:
		a.log.LogWarnf("HandleImplicateRecoverReceived: unexpected msgImplicateRecover.kind=%v, message: %+v", msg.Payload.kind, msg)
		return nil
	}
}

func (a *ACSS) handleImplicateRecoverPending() []gpa.MessageOut {
	//
	// Only process the IMPLICATE/RECOVER messages, if this node has RBC completed.
	if a.rbcOut == nil {
		return nil
	}
	postponedIRMsgs := []gpa.MessageIn[MsgImplicateRecover]{}
	var msgs []gpa.MessageOut
	for _, m := range a.pendingIRMsgs {
		switch m.Payload.kind {
		case msgImplicateRecoverKindIMPLICATE:
			// Only handle the IMPLICATE messages when output is already produced to implement the following:
			//
			// >     if out == true:
			// >       send <RECOVER, i, skᵢ> to all parties
			// >       return
			//
			if a.output {
				msgs = slices.Concat(msgs, a.handleImplicate(m))
			} else {
				postponedIRMsgs = append(postponedIRMsgs, m)
			}
		case msgImplicateRecoverKindRECOVER:
			msgs = slices.Concat(msgs, a.handleRecover(m))
		default:
			a.log.LogWarnf("HandleImplicateRecoverReceived: unexpected msgImplicateRecover.kind=%v, message: %+v", m.Payload.kind, m)
			// Don't return here, we are just dropping incorrect message.
		}
	}
	a.pendingIRMsgs = postponedIRMsgs
	return msgs
}

// Here the RBC is assumed to be completed already, OUT is set and the private key is checked.
//
// > on receiving <IMPLICATE, j, skⱼ>:
// >   sⱼ := PKI.Dec(eⱼ, skⱼ)
// >   if decrypt fails or VSS.Verify(C, j, sⱼ) == false:
// >     if out == true:
// >       send <RECOVER, i, skᵢ> to all parties
// >       return
//
// NOTE: We assume `if out == true:` stands for a wait for such condition.
func (a *ACSS) handleImplicate(msg gpa.MessageIn[MsgImplicateRecover]) []gpa.MessageOut {
	peerIndex := a.peerIndex(msg.Sender)
	if peerIndex == -1 {
		a.log.LogWarnf("implicate received from unknown peer: %v", msg.Sender)
		return nil
	}
	//
	// Check message duplicates.
	if _, ok := a.implicateRecv[msg.Sender]; ok {
		// Received the implicate before, just ignore it.
		return nil
	}
	a.implicateRecv[msg.Sender] = true
	//
	// Check implicate.
	secret, err := crypto.CheckImplicate(a.suite, a.rbcOut.PubKey, a.peerPKs[msg.Sender], msg.Payload.data)
	if err != nil {
		a.log.LogWarnf("Invalid implication received: %v", err)
		return nil
	}
	_, err = crypto.DecryptShare(a.suite, a.rbcOut, peerIndex, secret)
	if err == nil {
		// if we are able to decrypt the share, the implication is not correct
		a.log.LogWarn("encrypted share is valid")
		return nil
	}
	//
	// Create the reveal message.
	return a.broadcastRecover()
}

// Here the RBC is assumed to be completed already and the private key is checked.
//
// >     on receiving <RECOVER, j, skⱼ>:
// >       sⱼ := PKI.Dec(eⱼ, skⱼ)
// >       if VSS.Verify(C, j, sⱼ): T = T ∪ {sⱼ}
// >
// >     wait until len(T) >= f+1:
// >       sᵢ = SSS.Recover(T, f+1, n)(i)
// >       out = true
// >       output sᵢ
func (a *ACSS) handleRecover(msg gpa.MessageIn[MsgImplicateRecover]) []gpa.MessageOut {
	if a.output {
		// Ignore the RECOVER messages, if we are done with the output.
		return nil
	}
	peerIndex := a.peerIndex(msg.Sender)
	if peerIndex == -1 {
		a.log.LogWarnf("Recover received from unexpected sender: %v", msg.Sender)
		return nil
	}
	if _, ok := a.recoverRecv[msg.Sender]; ok {
		a.log.LogWarnf("Recover was already received from %v", msg.Sender)
		return nil
	}

	peerSecret, err := crypto.DecryptShare(a.suite, a.rbcOut, peerIndex, msg.Payload.data)
	if err != nil {
		a.log.LogWarn("invalid secret revealed")
		return nil
	}
	a.recoverRecv[msg.Sender] = peerSecret

	// >     wait until len(T) >= f+1:
	// >       sᵢ = SSS.Recover(T, f+1, n)(i)
	// >       out = true
	// >       output sᵢ
	if len(a.recoverRecv) >= a.f+1 {
		priShares := []*share.PriShare{}
		for i := range a.recoverRecv {
			priShares = append(priShares, a.recoverRecv[i])
		}

		myPriShare, err := crypto.InterpolateShare(a.suite, priShares, a.n, a.myIdx)
		if err != nil {
			a.log.LogWarnf("Failed to recover pri-poly: %v", err)
		}
		a.outS = myPriShare
		a.output = true
		return nil
	}

	return nil
}

func (a *ACSS) broadcastVote(voteKind msgVoteKind) []gpa.MessageOut {
	return lo.Map(a.peerIdx, func(peer gpa.NodeID, _ int) gpa.MessageOut {
		return gpa.NewMessageOut(peer, MsgVote{
			kind: voteKind,
		})
	})
}

func (a *ACSS) broadcastImplicate(reason error) []gpa.MessageOut {
	a.log.LogWarnf("Sending implicate because of: %v", reason)
	implicate := crypto.Implicate(a.suite, a.rbcOut.PubKey, a.mySK)
	return a.broadcastImplicateRecover(msgImplicateRecoverKindIMPLICATE, implicate)
}

func (a *ACSS) broadcastRecover() []gpa.MessageOut {
	secret := crypto.Secret(a.suite, a.rbcOut.PubKey, a.mySK)
	return a.broadcastImplicateRecover(msgImplicateRecoverKindRECOVER, secret)
}

func (a *ACSS) broadcastImplicateRecover(kind msgImplicateKind, data []byte) []gpa.MessageOut {
	return lo.Map(a.peerIdx, func(peer gpa.NodeID, _ int) gpa.MessageOut {
		return gpa.NewMessageOut(peer, MsgImplicateRecover{
			kind: kind,
			i:    a.myIdx,
			data: data,
		})
	})
}

func (a *ACSS) tryOutput() {
	count := len(a.voteREADYRecv)
	if count >= (a.n-a.f) && a.outS != nil {
		a.output = true
	}
}

func (a *ACSS) peerIndex(peer gpa.NodeID) int {
	for i := range a.peerIdx {
		if a.peerIdx[i] == peer {
			return i
		}
	}
	return -1
}

func (a *ACSS) Output() gpa.Output {
	if a.output {
		return &Output{
			PriShare: a.outS,
			Commits:  a.rbcOut.Commits,
		}
	}
	return nil
}

func (a *ACSS) StatusString() string {
	return fmt.Sprintf("{ACSS, output=%v, rbc=%v}", a.output, a.rbc.StatusString())
}
