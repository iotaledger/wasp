// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// package acss implements "Asynchronous Complete Secret Sharing" as described in
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
//
package acss

import (
	"fmt"

	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/gpa/acss/crypto"
	rbc "github.com/iotaledger/wasp/packages/gpa/rbc/bracha"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/suites"
	"golang.org/x/xerrors"
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
	log           *logger.Logger
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
	log *logger.Logger, // A logger to use.
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
		rbc:           rbc.New(peers, f, me, dealer, func(b []byte) bool { return true }),
		rbcOut:        nil, // Will be set on output from the RBC.
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
	if a.myIdx = a.peerIndex(me); a.myIdx == -1 {
		panic("i'm not in the peer list")
	}
	return gpa.NewOwnHandler(me, &a)
}

//
// Input for the algorithm is the secret to share.
// It can be provided by the dealer only.
//
func (a *acssImpl) Input(input gpa.Input) []gpa.Message {
	if a.me != a.dealer {
		panic(xerrors.Errorf("only dealer can initiate the sharing"))
	}
	if input == nil {
		panic(xerrors.Errorf("we expect kyber.Scalar as input"))
	}
	return a.handleInput(input.(kyber.Scalar))
}

//
// Receive all the messages and route them to the appropriate handlers.
//
func (a *acssImpl) Message(msg gpa.Message) []gpa.Message {
	switch m := msg.(type) {
	case *msgWrapper:
		switch m.subsystem {
		case msgWrapperRBC:
			return a.handleRBCMessage(m)
		default:
			panic(xerrors.Errorf("unexpected wrapped message: %+v", m))
		}
	case *msgRBCCEOutput:
		return a.handleRBCOutput(m)
	case *msgVote:
		switch m.kind {
		case msgVoteOK:
			return a.handleVoteOK(m)
		case msgVoteREADY:
			return a.handleVoteREADY(m)
		default:
			panic(xerrors.Errorf("unexpected vote message: %+v", m))
		}
	case *msgImplicateRecover:
		return a.handleImplicateRecoverReceived(m)
	default:
		panic(xerrors.Errorf("unexpected message: %+v", msg))
	}
}

//
// > // dealer with input s
// > sample random polynomial ϕ such that ϕ(0) = s
// > C, S := VSS.Share(ϕ, f+1, n)
// > E := [PKI.Enc(S[i], pkᵢ) for each party i]
// >
// > // party i (including the dealer)
// > RBC(C||E)
//
func (a *acssImpl) handleInput(secretToShare kyber.Scalar) []gpa.Message {
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
	rbcCEPayloadBytes, err := (&msgRBCCEPayload{suite: a.suite, data: data}).MarshalBinary()
	if err != nil {
		panic(xerrors.Errorf("cannot serialize msg_rbc_ce: %w", err))
	}
	return WrapMessages(msgWrapperRBC, a.rbc.Input(rbcCEPayloadBytes))
}

//
// Delegate received messages to the RBC and handle its output.
//
// > // party i (including the dealer)
// > RBC(C||E)
//
func (a *acssImpl) handleRBCMessage(m *msgWrapper) []gpa.Message {
	wasOut := a.rbc.Output() != nil // To send the msgRBCCEOutput message once (for perf reasons).
	msgs := WrapMessages(msgWrapperRBC, a.rbc.Message(m.wrapped))
	if out := a.rbc.Output(); !wasOut && out != nil {
		// Send the result for self as a message (maybe the code will look nicer this way).
		outParsed := &msgRBCCEPayload{suite: a.suite}
		if err := outParsed.UnmarshalBinary(out.([]byte)); err != nil {
			panic(xerrors.Errorf("cannot unmarshal msgRBCCEPayload"))
		}
		msgs = append(msgs, &msgRBCCEOutput{me: a.me, payload: outParsed})
	}
	return msgs
}

//
// Upon receiving the RBC output...
//
// > sᵢ := PKI.Dec(eᵢ, skᵢ)
// > if decrypt fails or VSS.Verify(C, i, sᵢ) == false:
// >   send <IMPLICATE, i, skᵢ> to all parties
// > else:
// >   send <OK>
//
func (a *acssImpl) handleRBCOutput(m *msgRBCCEOutput) []gpa.Message {
	if a.outS != nil || a.rbcOut != nil {
		// Take the first RBC output only.
		return gpa.NoMessages()
	}
	a.log.Debugf("handleRBCOutput: %+v", m)
	//
	// Store the broadcast result and process pending IMPLICATE/RECOVER messages, if any.
	deal, err := crypto.DealUnmarshalBinary(a.suite, a.n, m.payload.data)
	if err != nil {
		panic(xerrors.Errorf("cannot unmarshal msgRBCCEPayload.data"))
	}
	a.rbcOut = deal
	msgs := a.handleImplicateRecoverPending(gpa.NoMessages())
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

//
// > on receiving <OK> from n-f parties:
// >   send <READY> to all parties
//
func (a *acssImpl) handleVoteOK(m *msgVote) []gpa.Message {
	a.log.Debugf("handleVoteOK: %+v", m)
	a.voteOKRecv[m.sender] = true
	count := len(a.voteOKRecv)
	if !a.voteREADYSent && count >= (a.n-a.f) {
		a.voteREADYSent = true
		return a.broadcastVote(msgVoteREADY, gpa.NoMessages())
	}
	return gpa.NoMessages()
}

//
// > on receiving <READY> from f+1 parties:
// >   send <READY> to all parties
// >
// > on receiving <READY> from n-f parties:
// >   if sᵢ is valid:
// >     out = true
// >     output sᵢ
//
func (a *acssImpl) handleVoteREADY(m *msgVote) []gpa.Message {
	a.voteREADYRecv[m.sender] = true
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
func (a *acssImpl) handleImplicateRecoverReceived(m *msgImplicateRecover) []gpa.Message {
	if a.rbcOut == nil {
		a.pendingIRMsgs = append(a.pendingIRMsgs, m)
		return gpa.NoMessages()
	}
	switch m.kind {
	case msgImplicateRecoverKindIMPLICATE:
		return a.handleImplicate(m)
	case msgImplicateRecoverKindRECOVER:
		return a.handleRecover(m)
	default:
		panic(xerrors.Errorf("handleImplicateRecoverReceived: unexpected msgImplicateRecover.kind=%v, message: %+v", m.kind, m))
	}
}

func (a *acssImpl) handleImplicateRecoverPending(msgs []gpa.Message) []gpa.Message {
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
				msgs = append(msgs, a.handleImplicate(m)...)
			} else {
				postponedIRMsgs = append(postponedIRMsgs, m)
			}
		case msgImplicateRecoverKindRECOVER:
			msgs = append(msgs, a.handleRecover(m)...)
		default:
			panic(xerrors.Errorf("handleImplicateRecoverReceived: unexpected msgImplicateRecover.kind=%v, message: %+v", m.kind, m))
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
//
func (a *acssImpl) handleImplicate(m *msgImplicateRecover) []gpa.Message {
	peerIndex := a.peerIndex(m.sender)
	if peerIndex == -1 {
		a.log.Warnf("implicate received from unknown peer: %v", m.sender)
		return gpa.NoMessages()
	}
	//
	// Check message duplicates.
	if _, ok := a.implicateRecv[m.sender]; ok {
		// Received the implicate before, just ignore it.
		return gpa.NoMessages()
	}
	a.implicateRecv[m.sender] = true
	//
	// Check implicate.
	secret, err := crypto.CheckImplicate(a.suite, a.rbcOut.PubKey, a.peerPKs[m.sender], m.data)
	if err != nil {
		a.log.Warnf("Invalid implication received: %v", err)
		return gpa.NoMessages()
	}
	_, err = crypto.DecryptShare(a.suite, a.rbcOut, peerIndex, secret)
	if err == nil {
		// if we are able to decrypt the share, the implication is not correct
		a.log.Warnf("encrypted share is valid")
		return gpa.NoMessages()
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
//
func (a *acssImpl) handleRecover(m *msgImplicateRecover) []gpa.Message {
	if a.output {
		// Ignore the RECOVER messages, if we are done with the output.
		return gpa.NoMessages()
	}
	peerIndex := a.peerIndex(m.sender)
	if peerIndex == -1 {
		a.log.Warnf("Recover received from unexpected sender: %v", m.sender)
		return gpa.NoMessages()
	}
	if _, ok := a.recoverRecv[m.sender]; ok {
		a.log.Warnf("Recover was already received from %v", m.sender)
		return gpa.NoMessages()
	}

	peerSecret, err := crypto.DecryptShare(a.suite, a.rbcOut, peerIndex, m.data)
	if err != nil {
		a.log.Warnf("invalid secret revealed")
		return gpa.NoMessages()
	}
	a.recoverRecv[m.sender] = peerSecret

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
			a.log.Warnf("Failed to recover pri-poly: %v", err)
		}
		a.outS = myPriShare
		a.output = true
		return gpa.NoMessages()
	}

	return gpa.NoMessages()
}

func (a *acssImpl) broadcastVote(voteKind msgVoteKind, msgs []gpa.Message) []gpa.Message {
	for i := range a.peerIdx {
		msgs = append(msgs, &msgVote{sender: a.me, recipient: a.peerIdx[i], kind: voteKind})
	}
	return msgs
}

func (a *acssImpl) broadcastImplicate(reason error, msgs []gpa.Message) []gpa.Message {
	a.log.Warnf("Sending implicate because of: %v", reason)
	implicate := crypto.Implicate(a.suite, a.rbcOut.PubKey, a.mySK)
	return a.broadcastImplicateRecover(msgImplicateRecoverKindIMPLICATE, implicate, msgs)
}

func (a *acssImpl) broadcastRecover(msgs []gpa.Message) []gpa.Message {
	secret := crypto.Secret(a.suite, a.rbcOut.PubKey, a.mySK)
	return a.broadcastImplicateRecover(msgImplicateRecoverKindRECOVER, secret, msgs)
}

func (a *acssImpl) broadcastImplicateRecover(kind msgImplicateKind, data []byte, msgs []gpa.Message) []gpa.Message {
	for i := range a.peerIdx {
		msgs = append(msgs, &msgImplicateRecover{kind: kind, recipient: a.peerIdx[i], i: a.myIdx, data: data})
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
