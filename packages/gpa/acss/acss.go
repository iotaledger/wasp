// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// package acss implements "Asynchronous Complete Secret Sharing" as described in
//
//	https://iotaledger.github.io/crypto-tss/talks/async-dkg/slides-async-dkg.html#/5/6
//
// Here is a copy of the pseudocode from the slide mentioned above (just in case):
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
// More details and references to the papers are bellow:
//
// Here the references for the Asynchronous Secret-Sharing that I was referring to.
// It is purely based on (Feldman) Verifiable Secret Sharing and does not rely on any PVSS schemes
// requiring fancy NIZKP (and thus trades network-complexity vs computational-complexity):
//
//   * [1], Section IV. A. we use the ACSS scheme from [2] but replace its Pedersen
//     commitment with a Feldman polynomial commitment to achieve Homomorphic-Partial-Commitment.
//
//   * In [2], Section 5.3. they explain the Pedersen-based hbACSS0 and give some proof sketch.
//     The complete description and analysis of hbACSS0 can be found in [3]. However, as mentioned
//     before they use Kate-commitments instead of Feldman/Pedersen. This has better message
//     complexity especially when multiple secrets are shared at the same time, but in our case
//     that would need to be replaced with Feldman making it much simpler and not losing any security.
//     Actually, [3] is just a pre-print, the official published version is [4], but [4] also contains
//     other, non-relevant, variants like hbACSS1 and hbACSS2 and much more analysis.
//     So, I found [3] a bit more helpful, although it is just the preliminary version.
//     They also provide their reference implementation in [5], which is also what the
//     authors of [1] used for their practical DKG results.
//
// [1] Practical Asynchronous Distributed Key Generation https://eprint.iacr.org/2021/1591
// [2] Asynchronous Data Dissemination and its Applications https://eprint.iacr.org/2021/777
// [3] Brief Note: Asynchronous Verifiable Secret Sharing with Optimal Resilience and Linear Amortized Overhead https://arxiv.org/pdf/1902.06095.pdf
// [4] hbACSS: How to Robustly Share Many Secrets https://eprint.iacr.org/2021/159
// [5] https://github.com/tyurek/hbACSS
//
// A PoC implementation: <https://github.com/Wollac/async.go>
//
package acss

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/gpa"
	rbc "github.com/iotaledger/wasp/packages/gpa/rbc/bracha"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/encrypt/ecies"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/suites"
	"golang.org/x/xerrors"
)

type acssImpl struct {
	suite         suites.Suite
	n             int
	f             int
	me            gpa.NodeID
	mySK          kyber.Scalar
	myPK          kyber.Point
	myIdx         int
	dealer        gpa.NodeID                 // A node that is recognized as a dealer.
	dealCB        func(int, []byte) []byte   // Callback to be called on the encrypted deals (for tests actually).
	peerPKs       map[gpa.NodeID]kyber.Point // Peer public keys.
	peerIdx       []gpa.NodeID               // Particular order of the nodes (position in the polynomial).
	rbc           gpa.GPA                    // RBC to share `C||E`.
	rbcOutC       *share.PubPoly             // C -- Commitment from the dealer.
	rbcOutE       [][]byte                   // E -- Encrypted Private Shares.
	voteOKRecv    map[gpa.NodeID]bool        // A set of received OK votes.
	voteREADYRecv map[gpa.NodeID]bool        // A set of received READY votes.
	voteREADYSent bool                       // Have we sent our READY vote?
	pendingIRMsgs []*msgImplicateRecover     // I/R messages are buffered, if the RBC is not completed yet.
	recoverT      map[int]*share.PriShare    // Private shares from the RECOVER messages.
	outS          *share.PriShare            // Our share of the secret (decrypted from rbcOutE).
	output        bool
	log           *logger.Logger
}

//
// NOTE: The secret key `mySK` have to be temporary, as it is revealed in the case
// when the dealer is detected to be faulty and the IMPLICATE/RECOVER procedure is used.
//
func New(
	suite suites.Suite,
	peers []gpa.NodeID,
	peerPKs map[gpa.NodeID]kyber.Point,
	f int,
	me gpa.NodeID,
	mySK kyber.Scalar,
	dealer gpa.NodeID,
	dealCB func(int, []byte) []byte,
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
		myIdx:         myIdx,
		dealer:        dealer,
		dealCB:        dealCB,
		peerPKs:       peerPKs,
		peerIdx:       peers,
		rbc:           rbc.New(peers, f, me, dealer, func(b []byte) bool { return true }),
		rbcOutC:       nil, // Will be set on output from the RBC.
		rbcOutE:       nil, // Will be set on output from the RBC.
		voteOKRecv:    map[gpa.NodeID]bool{},
		voteREADYRecv: map[gpa.NodeID]bool{},
		voteREADYSent: false,
		pendingIRMsgs: []*msgImplicateRecover{},
		recoverT:      map[int]*share.PriShare{},
		outS:          nil,
		output:        false,
		log:           log,
	}
	return gpa.NewOwnHandler(me, &a)
}

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
	// > sample random polynomial ϕ such that ϕ(0) = s
	priPoly := share.NewPriPoly(a.suite, a.f+1, secretToShare, a.suite.RandomStream())

	// > C, S := VSS.Share(ϕ, f+1, n)
	C := priPoly.Commit(nil)
	S := priPoly.Shares(a.n)

	// > E := [PKI.Enc(S[i], pkᵢ) for each party i]
	E := make([][]byte, a.n)
	for i, peerID := range a.peerIdx {
		if i != S[i].I {
			panic("i != S[i].I")
		}
		Si, err := S[S[i].I].V.MarshalBinary()
		if err != nil {
			panic(xerrors.Errorf("cannot serialize share: %w", err))
		}
		Ei, err := ecies.Encrypt(a.suite, a.peerPKs[peerID], Si, a.suite.Hash)
		if err != nil {
			panic(xerrors.Errorf("cannot encrypt share: %w", err))
		}
		E[i] = a.dealCB(i, Ei)
	}

	// > RBC(C||E)
	rbcCEPayloadBytes, err := (&msgRBCCEPayload{suite: a.suite, C: C, E: E}).MarshalBinary()
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
	msgs := WrapMessages(msgWrapperRBC, a.rbc.Message(m.wrapped))
	if out := a.rbc.Output(); out != nil {
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
	if a.outS != nil {
		// Take the first RBC output only.
		return gpa.NoMessages()
	}
	a.log.Debugf("handleRBCOutput: %+v", m)
	//
	// Store the broadcast result and process pending IMPLICATE/RECOVER messages, if any.
	a.rbcOutC = m.payload.C
	a.rbcOutE = m.payload.E
	msgs := a.handleImplicateRecoverPending(gpa.NoMessages())
	//
	// Process the RBC output, as described above.
	myShare, err := a.tryDecryptVerifyPriShare(a.myIdx, a.mySK)
	if err != nil {
		return a.broadcastImplicate(err, msgs)
	}
	a.outS = myShare
	a.tryOutput() // Maybe the READY messages are already received.
	return a.broadcastVote(msgVoteOK, msgs)
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
	a.log.Debugf("handleVoteREADY: %+v", m)
	a.voteREADYRecv[m.sender] = true
	count := len(a.voteREADYRecv)
	msgs := gpa.NoMessages()
	if !a.voteREADYSent && count >= (a.f+1) {
		msgs = a.broadcastVote(msgVoteREADY, msgs)
		a.voteREADYSent = true
	}
	a.tryOutput()
	return msgs
}

// It is possible that we are receiving IMPLICATE/RECOVER messages before our RBC is completed.
// We store these messages for processing after that, if RBC is not done and process it otherwise.
func (a *acssImpl) handleImplicateRecoverReceived(m *msgImplicateRecover) []gpa.Message {
	if !a.checkPrivateKey(m.i, m.sk) {
		a.log.Warnf("handleImplicateRecoverReceived: node[%v]=%v provided invalid secret key, will ignore the message.", m.i, a.peerIdx[m.i])
		return gpa.NoMessages()
	}
	if a.rbcOutC == nil && a.rbcOutE == nil {
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
	if a.rbcOutC == nil && a.rbcOutE == nil {
		return msgs
	}
	for _, m := range a.pendingIRMsgs {
		switch m.kind {
		case msgImplicateRecoverKindIMPLICATE:
			msgs = append(msgs, a.handleImplicate(m)...)
		case msgImplicateRecoverKindRECOVER:
			msgs = append(msgs, a.handleRecover(m)...)
		default:
			panic(xerrors.Errorf("handleImplicateRecoverReceived: unexpected msgImplicateRecover.kind=%v, message: %+v", m.kind, m))
		}
	}
	a.pendingIRMsgs = []*msgImplicateRecover{}
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
// TODO: The check `if out == true:` is considered a wait??? For now we assume yes.
//
func (a *acssImpl) handleImplicate(m *msgImplicateRecover) []gpa.Message {
	_, err := a.tryDecryptVerifyPriShare(m.i, m.sk)
	if err != nil {
		return a.broadcastRecover(gpa.NoMessages())
	}
	return gpa.NoMessages()
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

	// >     on receiving <RECOVER, j, skⱼ>:
	// >       sⱼ := PKI.Dec(eⱼ, skⱼ)
	// >       if VSS.Verify(C, j, sⱼ): T = T ∪ {sⱼ}
	sJ, err := a.tryDecryptVerifyPriShare(m.i, m.sk)
	if err != nil {
		a.log.Warnf("recover message cannot be used to decrypt share: %v", err)
		return gpa.NoMessages()
	}
	a.recoverT[m.i] = sJ

	// >     wait until len(T) >= f+1:
	// >       sᵢ = SSS.Recover(T, f+1, n)(i)
	// >       out = true
	// >       output sᵢ
	if len(a.recoverT) >= a.f+1 {
		priShares := []*share.PriShare{}
		for i := range a.recoverT {
			priShares = append(priShares, a.recoverT[i])
		}
		priPoly, err := share.RecoverPriPoly(a.suite, priShares, a.f+1, a.n)
		if err != nil {
			a.log.Warnf("Failed to recover pri-poly: %v", err)
		}
		a.outS = priPoly.Shares(a.n)[a.myIdx]
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
	return a.broadcastImplicateRecover(msgImplicateRecoverKindIMPLICATE, msgs)
}

func (a *acssImpl) broadcastRecover(msgs []gpa.Message) []gpa.Message {
	return a.broadcastImplicateRecover(msgImplicateRecoverKindRECOVER, msgs)
}

func (a *acssImpl) broadcastImplicateRecover(kind msgImplicateKind, msgs []gpa.Message) []gpa.Message {
	for i := range a.peerIdx {
		msgs = append(msgs, &msgImplicateRecover{kind: kind, recipient: a.peerIdx[i], i: a.myIdx, sk: a.mySK})
	}
	return msgs
}

//
// Assume rbcOutE and rbcOutE are already set.
//
func (a *acssImpl) tryDecryptVerifyPriShare(j int, skJ kyber.Scalar) (*share.PriShare, error) {
	decrypted, err := ecies.Decrypt(a.suite, skJ, a.rbcOutE[j], a.suite.Hash)
	if err != nil {
		return nil, xerrors.Errorf("failed to decrypt share: %v", err)
	}
	jShare := a.suite.Scalar()
	if err := jShare.UnmarshalBinary(decrypted); err != nil {
		return nil, xerrors.Errorf("failed to unmarshal share: %v", err)
	}
	jPriShare := &share.PriShare{I: j, V: jShare}
	if !a.rbcOutC.Check(jPriShare) {
		return nil, xerrors.Errorf("share verification failed")
	}
	return jPriShare, nil
}

func (a *acssImpl) tryOutput() {
	count := len(a.voteREADYRecv)
	if count >= (a.n-a.f) && a.outS != nil {
		a.output = true
	}
}

func (a *acssImpl) checkPrivateKey(j int, skJ kyber.Scalar) bool {
	return a.peerPKs[a.peerIdx[j]].Equal(a.suite.Point().Mul(skJ, nil))
}

func (a *acssImpl) Output() gpa.Output {
	if a.output {
		return a.outS
	}
	return nil
}
