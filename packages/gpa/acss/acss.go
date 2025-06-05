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

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/suites"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/gpa/acss/crypto"
	rbc "github.com/iotaledger/wasp/packages/gpa/rbc/bracha"
)

const (
	subsystemRBC byte = iota
)

type Output struct {
	PriShare *share.PriShare // Private share, received by this instance.
	Commits  []kyber.Point   // Feldman's commitment to the shared polynomial.
}

type acssImpl struct {
	suite         suites.Suite
	n             int
	f             int
	me            gpa.NodeID
	mySK          kyber.Scalar
	myPK          kyber.Point
	myIdx         int
	dealer        gpa.NodeID                     // A node that is recognized as a dealer.
	dealCB        func(int, []byte) []byte       // Callback to be called on the encrypted deals (for tests actually).
	peerPKs       map[gpa.NodeID]kyber.Point     // Peer public keys.
	peerIdx       []gpa.NodeID                   // Particular order of the nodes (position in the polynomial).
	rbc           gpa.GPA                        // RBC to share `C||E`.
	rbcOut        *crypto.Deal                   // Deal broadcasted by the dealer.
	voteOKRecv    map[gpa.NodeID]bool            // A set of received OK votes.
	voteREADYRecv map[gpa.NodeID]bool            // A set of received READY votes.
	voteREADYSent bool                           // Have we sent our READY vote?
	pendingIRMsgs []*msgImplicateRecover         // I/R messages are buffered, if the RBC is not completed yet.
	implicateRecv map[gpa.NodeID]bool            // To check, that implicate only received once from a node.
	recoverRecv   map[gpa.NodeID]*share.PriShare // Private shares from the RECOVER messages.
	outS          *share.PriShare                // Our share of the secret (decrypted from rbcOutE).
	output        bool
	msgWrapper    *gpa.MsgWrapper
	log           log.Logger
}

var _ gpa.GPA = &acssImpl{}

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
) gpa.GPA {
	n := len(peers)
	if dealCB == nil {
		dealCB = func(i int, b []byte) []byte { return b }
	}
	a := acssImpl{
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
		pendingIRMsgs: []*msgImplicateRecover{},
		implicateRecv: map[gpa.NodeID]bool{},
		recoverRecv:   map[gpa.NodeID]*share.PriShare{},
		outS:          nil,
		output:        false,
		log:           log,
	}
	a.msgWrapper = gpa.NewMsgWrapper(msgTypeWrapped, func(subsystem byte, index int) (gpa.GPA, error) {
		if subsystem == subsystemRBC {
			if index != 0 {
				return nil, fmt.Errorf("unknown rbc index: %v", index)
			}
			return a.rbc, nil
		}
		return nil, fmt.Errorf("unknown subsystem: %v", subsystem)
	})
	if a.myIdx = a.peerIndex(me); a.myIdx == -1 {
		panic("i'm not in the peer list")
	}
	return gpa.NewOwnHandler(me, &a)
}

// Input for the algorithm is the secret to share.
// It can be provided by the dealer only.
func (a *acssImpl) Input(input gpa.Input) gpa.OutMessages {
	if a.me != a.dealer {
		panic(errors.New("only dealer can initiate the sharing"))
	}
	if input == nil {
		panic(errors.New("we expect kyber.Scalar as input"))
	}
	return a.handleInput(input.(kyber.Scalar))
}

// Receive all the messages and route them to the appropriate handlers.
func (a *acssImpl) Message(msg gpa.Message) gpa.OutMessages {
	switch m := msg.(type) {
	case *gpa.WrappingMsg:
		switch m.Subsystem() {
		case subsystemRBC:
			return a.handleRBCMessage(m)
		default:
			a.log.LogWarnf("unexpected wrapped message subsystem: %+v", m)
			return nil
		}
	case *msgVote:
		switch m.kind {
		case msgVoteOK:
			return a.handleVoteOK(m)
		case msgVoteREADY:
			return a.handleVoteREADY(m)
		default:
			a.log.LogWarnf("unexpected vote message: %+v", m)
			return nil
		}
	case *msgImplicateRecover:
		return a.handleImplicateRecoverReceived(m)
	default:
		panic(fmt.Errorf("unexpected message: %+v", msg))
	}
}

// > // dealer with input s
// > sample random polynomial ϕ such that ϕ(0) = s
// > C, S := VSS.Share(ϕ, f+1, n)
// > E := [PKI.Enc(S[i], pkᵢ) for each party i]
// >
// > // party i (including the dealer)
// > RBC(C||E)
func (a *acssImpl) handleInput(secretToShare kyber.Scalar) gpa.OutMessages {
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
	rbcCEPayloadBytes := bcs.MustMarshal(&msgRBCCEPayload{suite: a.suite, data: data})
	msgs := a.msgWrapper.WrapMessages(subsystemRBC, 0, a.rbc.Input(rbcCEPayloadBytes))
	return a.tryHandleRBCTermination(false, msgs)
}

// Delegate received messages to the RBC and handle its output.
//
// > // party i (including the dealer)
// > RBC(C||E)
func (a *acssImpl) handleRBCMessage(m *gpa.WrappingMsg) gpa.OutMessages {
	wasOut := a.rbc.Output() != nil // To send the msgRBCCEOutput message once (for perf reasons).
	msgs := a.msgWrapper.WrapMessages(subsystemRBC, 0, a.rbc.Message(m.Wrapped()))
	return a.tryHandleRBCTermination(wasOut, msgs)
}

func (a *acssImpl) tryHandleRBCTermination(wasOut bool, msgs gpa.OutMessages) gpa.OutMessages {
	if out := a.rbc.Output(); !wasOut && out != nil {
		// Send the result for self as a message (maybe the code will look nicer this way).
		outParsed, err := bcs.UnmarshalInto(out.([]byte), &msgRBCCEPayload{suite: a.suite})
		if err != nil {
			outParsed = &msgRBCCEPayload{err: err}
		}
		msgs.AddAll(a.handleRBCOutput(outParsed))
	}
	return msgs
}

// Upon receiving the RBC output...
//
// > sᵢ := PKI.Dec(eᵢ, skᵢ)
// > if decrypt fails or VSS.Verify(C, i, sᵢ) == false:
// >   send <IMPLICATE, i, skᵢ> to all parties
// > else:
// >   send <OK>
func (a *acssImpl) handleRBCOutput(rbcOutput *msgRBCCEPayload) gpa.OutMessages {
	if a.outS != nil || a.rbcOut != nil {
		// Take the first RBC output only.
		return nil
	}
	msgs := gpa.NoMessages()
	//
	// Store the broadcast result and process pending IMPLICATE/RECOVER messages, if any.
	if rbcOutput.err != nil {
		return a.broadcastImplicate(rbcOutput.err, msgs)
	}
	deal, err := crypto.DealUnmarshalBinary(a.suite, a.n, rbcOutput.data)
	if err != nil {
		return a.broadcastImplicate(errors.New("cannot unmarshal msgRBCCEPayload.data"), msgs)
	}
	a.rbcOut = deal
	msgs = a.handleImplicateRecoverPending(msgs)
	//
	// Process the RBC output, as described above.
	secret := crypto.Secret(a.suite, a.rbcOut.PubKey, a.mySK)
	myShare, err := crypto.DecryptShare(a.suite, a.rbcOut, a.myIdx, secret)
	if err != nil {
		return a.broadcastImplicate(err, msgs)
	}
	a.outS = myShare
	a.tryOutput() // Maybe the READY messages are already received.
	return a.handleImplicateRecoverPending(a.broadcastVote(msgVoteOK, msgs))
}

// > on receiving <OK> from n-f parties:
// >   send <READY> to all parties
func (a *acssImpl) handleVoteOK(msg *msgVote) gpa.OutMessages {
	a.voteOKRecv[msg.Sender()] = true
	count := len(a.voteOKRecv)
	if !a.voteREADYSent && count >= (a.n-a.f) {
		a.voteREADYSent = true
		return a.broadcastVote(msgVoteREADY, gpa.NoMessages())
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
func (a *acssImpl) handleVoteREADY(msg *msgVote) gpa.OutMessages {
	a.voteREADYRecv[msg.Sender()] = true
	count := len(a.voteREADYRecv)
	msgs := gpa.NoMessages()
	if !a.voteREADYSent && count >= (a.f+1) {
		msgs = a.broadcastVote(msgVoteREADY, msgs)
		a.voteREADYSent = true
	}
	a.tryOutput()
	return a.handleImplicateRecoverPending(msgs)
}

// It is possible that we are receiving IMPLICATE/RECOVER messages before our RBC is completed.
// We store these messages for processing after that, if RBC is not done and process it otherwise.
func (a *acssImpl) handleImplicateRecoverReceived(msg *msgImplicateRecover) gpa.OutMessages {
	if a.rbcOut == nil {
		a.pendingIRMsgs = append(a.pendingIRMsgs, msg)
		return nil
	}
	switch msg.kind {
	case msgImplicateRecoverKindIMPLICATE:
		return a.handleImplicate(msg)
	case msgImplicateRecoverKindRECOVER:
		return a.handleRecover(msg)
	default:
		a.log.LogWarnf("handleImplicateRecoverReceived: unexpected msgImplicateRecover.kind=%v, message: %+v", msg.kind, msg)
		return nil
	}
}

func (a *acssImpl) handleImplicateRecoverPending(msgs gpa.OutMessages) gpa.OutMessages {
	//
	// Only process the IMPLICATE/RECOVER messages, if this node has RBC completed.
	if a.rbcOut == nil {
		return msgs
	}
	postponedIRMsgs := []*msgImplicateRecover{}
	for _, m := range a.pendingIRMsgs {
		switch m.kind {
		case msgImplicateRecoverKindIMPLICATE:
			// Only handle the IMPLICATE messages when output is already produced to implement the following:
			//
			// >     if out == true:
			// >       send <RECOVER, i, skᵢ> to all parties
			// >       return
			//
			if a.output {
				msgs.AddAll(a.handleImplicate(m))
			} else {
				postponedIRMsgs = append(postponedIRMsgs, m)
			}
		case msgImplicateRecoverKindRECOVER:
			msgs.AddAll(a.handleRecover(m))
		default:
			a.log.LogWarnf("handleImplicateRecoverReceived: unexpected msgImplicateRecover.kind=%v, message: %+v", m.kind, m)
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
func (a *acssImpl) handleImplicate(msg *msgImplicateRecover) gpa.OutMessages {
	peerIndex := a.peerIndex(msg.sender)
	if peerIndex == -1 {
		a.log.LogWarnf("implicate received from unknown peer: %v", msg.sender)
		return nil
	}
	//
	// Check message duplicates.
	if _, ok := a.implicateRecv[msg.sender]; ok {
		// Received the implicate before, just ignore it.
		return nil
	}
	a.implicateRecv[msg.sender] = true
	//
	// Check implicate.
	secret, err := crypto.CheckImplicate(a.suite, a.rbcOut.PubKey, a.peerPKs[msg.sender], msg.data)
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
	return a.broadcastRecover(gpa.NoMessages())
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
func (a *acssImpl) handleRecover(msg *msgImplicateRecover) gpa.OutMessages {
	if a.output {
		// Ignore the RECOVER messages, if we are done with the output.
		return nil
	}
	peerIndex := a.peerIndex(msg.sender)
	if peerIndex == -1 {
		a.log.LogWarnf("Recover received from unexpected sender: %v", msg.sender)
		return nil
	}
	if _, ok := a.recoverRecv[msg.sender]; ok {
		a.log.LogWarnf("Recover was already received from %v", msg.sender)
		return nil
	}

	peerSecret, err := crypto.DecryptShare(a.suite, a.rbcOut, peerIndex, msg.data)
	if err != nil {
		a.log.LogWarn("invalid secret revealed")
		return nil
	}
	a.recoverRecv[msg.sender] = peerSecret

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

func (a *acssImpl) broadcastVote(voteKind msgVoteKind, msgs gpa.OutMessages) gpa.OutMessages {
	for i := range a.peerIdx {
		msg := &msgVote{
			BasicMessage: gpa.NewBasicMessage(a.peerIdx[i]),
			kind:         voteKind,
		}
		msg.SetSender(a.me)
		msgs.Add(msg)
	}
	return msgs
}

func (a *acssImpl) broadcastImplicate(reason error, msgs gpa.OutMessages) gpa.OutMessages {
	a.log.LogWarnf("Sending implicate because of: %v", reason)
	implicate := crypto.Implicate(a.suite, a.rbcOut.PubKey, a.mySK)
	return a.broadcastImplicateRecover(msgImplicateRecoverKindIMPLICATE, implicate, msgs)
}

func (a *acssImpl) broadcastRecover(msgs gpa.OutMessages) gpa.OutMessages {
	secret := crypto.Secret(a.suite, a.rbcOut.PubKey, a.mySK)
	return a.broadcastImplicateRecover(msgImplicateRecoverKindRECOVER, secret, msgs)
}

func (a *acssImpl) broadcastImplicateRecover(kind msgImplicateKind, data []byte, msgs gpa.OutMessages) gpa.OutMessages {
	for i := range a.peerIdx {
		msgs.Add(&msgImplicateRecover{kind: kind, recipient: a.peerIdx[i], i: a.myIdx, data: data})
	}
	return msgs
}

func (a *acssImpl) tryOutput() {
	count := len(a.voteREADYRecv)
	if count >= (a.n-a.f) && a.outS != nil {
		a.output = true
	}
}

func (a *acssImpl) peerIndex(peer gpa.NodeID) int {
	for i := range a.peerIdx {
		if a.peerIdx[i] == peer {
			return i
		}
	}
	return -1
}

func (a *acssImpl) Output() gpa.Output {
	if a.output {
		return &Output{
			PriShare: a.outS,
			Commits:  a.rbcOut.Commits,
		}
	}
	return nil
}

func (a *acssImpl) StatusString() string {
	return fmt.Sprintf("{ACSS, output=%v, rbc=%v}", a.output, a.rbc.StatusString())
}
