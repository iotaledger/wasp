// Copyright 2020 IOTA Stiftung

// Package mostefaoui implements the Mostefaoui ABA consensus algorithm
// SPDX-License-Identifier: Apache-2.0
//
// Here we implement the Asynchronous Byzantine Binary Agreement by
// Mostefaoui et al., as described in the HBBFT paper:
//
// > Miller, A., Xia, Y., Croman, K., Shi, E., and Song, D. (2016). The Honey Badger of
// > BFT Protocols. In Proceedings of the 2016 ACM SIGSAC Conference on Computer
// > and Communications Security, CCS ’16, page 31–42, New York, NY, USA.
// > Association for Computing Machinery.
//
// The original paper by Mostefaoui is:
//
// > A. Mostefaoui, H. Moumen, and M. Raynal. Signature-free
// > asynchronous byzantine consensus with t< n/3 and o (n 2)
// > messages. In Proceedings of the 2014 ACM symposium on
// > Principles of distributed computing, pages 2–9. ACM, 2014.
//
// The HBBFT paper presents the algorithm as follows:
//
// > • upon receiving input b_input, set est_0 := b_input and proceed as
// >   follows in consecutive epochs, with increasing labels r:
// >     – multicast BVAL_r(est_r)
// >     – bin_values_r := {}
// >     – upon receiving BVAL_r(b) messages from f + 1 nodes, if
// >       BVAL_r(b) has not been sent, multicast BVAL_r(b)
// >     – upon receiving BVAL_r(b) messages from 2f + 1 nodes,
// >       bin_values_r := bin_values_r ∪ {b}
// >     – wait until bin_values_r != {}, then
// >         ∗ multicast AUX_r(w) where w ∈ bin_values_r
// >         ∗ wait until at least (N − f) AUX_r messages have been
// >           received, such that the set of values carried by these
// >           messages, vals are a subset of bin_values_r (note that
// >           bin_values_r may continue to change as BVAL_r messages
// >           are received, thus this condition may be triggered upon
// >           arrival of either an AUX_r or a BVAL_r message)
// >         ∗ s ← Coin_r.GetCoin()
// >         ∗ if vals = {b}, then
// >             · est_r+1 := b
// >             · if (b = s%2) then output b
// >         ∗ else est_r+1 := s%2
// > • continue looping until both a value b is output in some round r,
// >   and the value Coin_r' = b for some round r' > r.
//
// Additionally we add the STOP messages to make this algorithm terminating.
// The STOP messages are discussed in the original paper by Mostefaoui.
//
// This implementation is split to several parts to handle various rance
// conditions easier.
//
//   - varBinVals -- maintains the binValues variable and handles the BVAL messages.
//   - varAuxVals -- maintains the `vars` variable and handles the AUX messages.
//   - varDone -- tracks the termination condition for the algorithm.
//   - uponDecisionInputs -- a predicate waiting for the CC and AuxVals to be ready.
//
// All these parts are independent of each-other and are wired-up in this file.
// With these parts defined, the overall algorithm can be rephrased as follows:
//
// > • upon receiving input b_input, set est_0 := b_input and proceed as
// >   follows in consecutive epochs, with increasing labels r:
// >     - start the round r (for varBinVals, CC and others).
// >     - on each varBinVals update pass it to varAuxVals.
// >     - wait until varAuxVals != {} and s ← Coin_r.GetCoin()
// >         ∗ if vals = {b}, then
// >             · est_r+1 := b
// >             · if (b = s%2) then output b
// >         ∗ else est_r+1 := s%2
// > • continue looping varDone is true.
package mostefaoui

import (
	"fmt"
	"slices"

	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/gpa/cc/blssig"
	"github.com/iotaledger/wasp/v2/packages/gpa/cc/semi"
)

// Output is the structure provided as an output of the algorithm.
// If the value is undecided, untyped nil is returned.
// The Terminate field indicates, if this algorithm can be
// dropped (no other peers need any messages from this node).
type Output struct {
	Value      bool
	Terminated bool
}

// ABA is the public API for this protocol.
const (
	SubsystemID string = "aba"
)

type ABA struct {
	nodeIDs            []gpa.NodeID                 // Nodes in the consensus.
	nodeIdx            map[gpa.NodeID]bool          // For a fast check, if peer is known.
	round              int                          // The current round.
	varBinVals         *varBinVals                  // The `binValues` variable (based on BVAL msgs).
	varAuxVals         *varAuxVals                  // The `vals` variable (based on AUX msgs).
	varDone            *varDone                     // Termination condition.
	uponDecisionInputs *uponDecisionInputs          // Decision condition.
	ccInsts            []*semi.CCSemi               // Common coin instances for all the rounds.
	ccCreateFun        func(round int) *semi.CCSemi // Function to create CC instances.
	output             *Output                      // The current output of the algorithm.
	postponedMsgs      []gpa.MessageIn[MsgVote]     // Buffer for future round messages.
	log                log.Logger                   // A logger.
}

// New creates a single node for a consensus.
//
// Here `ccCreateFun` is used as a factory function to create Common Coin instances for each round.
// This way this implementation is made independent of particular CC instance. The created CC
// is expected to take `nil` inputs and produce `*bool` outputs.
func New(nodeIDs []gpa.NodeID, me gpa.NodeID, f int, ccCreateFun func(round int) *semi.CCSemi, log log.Logger) *ABA {
	nodeIdx := map[gpa.NodeID]bool{}
	for _, n := range nodeIDs {
		nodeIdx[n] = true
	}
	a := &ABA{
		nodeIDs:       nodeIDs,
		nodeIdx:       nodeIdx,
		round:         -1,
		ccInsts:       []*semi.CCSemi{},
		ccCreateFun:   ccCreateFun,
		output:        nil,
		postponedMsgs: []gpa.MessageIn[MsgVote]{},
		log:           log,
	}
	a.varBinVals = newBinVals(nodeIDs, f, a.uponBinValuesUpdated)
	a.varAuxVals = newAuxVals(nodeIDs, f, a.uponAuxValsReady)
	a.varDone = newVarDone(nodeIDs, me, f, a.uponTerminationCondition, log)
	a.uponDecisionInputs = newUponDecisionInputs(a.uponDecisionInputsReceived)
	return a
}

func (a *ABA) selectCC(round int) (*semi.CCSemi, error) {
	if round > a.round+10 {
		return nil, fmt.Errorf("cc round=%v to far in future, our round=%v", round, a.round)
	}
	return a.ccInst(round), nil
}

// Creates and returns a CC instance for a particular round.
// CC instances are not cleaned up, as the algorithm is supposed to terminate in few rounds.
func (a *ABA) ccInst(round int) *semi.CCSemi {
	if round >= len(a.ccInsts) {
		add := make([]*semi.CCSemi, round-len(a.ccInsts)+1)
		a.ccInsts = append(a.ccInsts, add...)
	}
	if a.ccInsts[round] == nil {
		a.ccInsts[round] = a.ccCreateFun(round)
	}
	return a.ccInsts[round]
}

// Input implements the gpa.GPA interface.
//
// > • upon receiving input b_input, set est_0 := b_input and proceed as
// >   follows in consecutive epochs, with increasing labels r:
func (a *ABA) Input(input gpa.Input) []gpa.MessageOut {
	if a.round != -1 {
		panic(fmt.Errorf("duplicate input to BBA: %v", input))
	}
	if _, ok := input.(bool); !ok {
		panic(fmt.Errorf("input for BBA has to be bool, received %T=%+v", input, input))
	}
	return a.startRound(0, input.(bool))
}

// Advances the algorithm to the next round.
//
// >     – multicast BVAL_r(est_r)
// >     – bin_values_r := {}
func (a *ABA) startRound(round int, est bool) []gpa.MessageOut {
	if a.output != nil && a.output.Terminated {
		// Don't start the next round if the algorithm is already terminated.
		return nil
	}
	if round != a.round+1 {
		panic(fmt.Errorf("non-sequential rounds %v->%v", a.round, round))
	}
	a.round = round
	a.varAuxVals.startRound(a.round)
	a.varDone.startRound(round)
	a.uponDecisionInputs.startRound()
	msgs := a.varBinVals.startRound(a.round, est)
	//
	// Start the CC.
	cc, err := a.selectCC(round)
	if err != nil {
		panic(fmt.Errorf("failed to select CC: %v", err))
	}

	msgs = slices.Concat(
		msgs,
		gpa.AddKey(SubsystemID, round, cc.Input(nil)),
	)

	if out := cc.Output(); out != nil {
		msgs = slices.Concat(msgs, a.uponDecisionInputs.ccOutputReceived(*out.(*bool)))
	}
	//
	// Resend postponed messages, if any.
	if len(a.postponedMsgs) > 0 {
		oldPostponedMsgs := a.postponedMsgs
		a.postponedMsgs = []gpa.MessageIn[MsgVote]{}
		for _, m := range oldPostponedMsgs {
			msgs = slices.Concat(msgs, a.HandleMsgVote(m))
		}
	}
	return msgs
}

func (a *ABA) HandleMsgVote(msgT gpa.MessageIn[MsgVote]) []gpa.MessageOut {
	if _, ok := a.nodeIdx[msgT.Sender]; !ok {
		a.log.LogWarnf("unknown sender: %+v", msgT)
		return nil // Unknown sender.
	}
	if msgT.Payload.round < a.round || (a.output != nil && a.output.Terminated) {
		return nil // Outdated message.
	}
	if msgT.Payload.round > a.round {
		a.postponedMsgs = append(a.postponedMsgs, msgT)
		return nil // Will be processed later.
	}
	switch msgT.Payload.voteType {
	case BVAL:
		return a.varBinVals.msgVoteBVALReceived(msgT)
	case AUX:
		return a.varAuxVals.msgVoteAUXReceived(msgT)
	}
	a.log.LogWarnf("unexpected msgVote message: %+v", msgT)
	return nil
}

func (a *ABA) HandleMsgDone(msgT gpa.MessageIn[MsgDone]) []gpa.MessageOut {
	if _, ok := a.nodeIdx[msgT.Sender]; !ok {
		return nil // Unknown sender.
	}
	return a.varDone.msgDoneReceived(msgT)
}

func (a *ABA) HandleCCMsg(ccIndex int, msgT gpa.MessageIn[blssig.MsgSigShare]) []gpa.MessageOut {
	cc, err := a.selectCC(ccIndex)
	if err != nil {
		a.log.LogWarnf("cannot select subsystem: %v", err)
		return nil
	}

	subMsgs := cc.HandleMsgSigShare(msgT)
	msgs := gpa.AddKey(SubsystemID, ccIndex, subMsgs)
	if ccIndex == a.round && !a.uponDecisionInputs.haveCC() {
		ccOut := cc.Output()
		if ccOut != nil {
			msgs = slices.Concat(msgs, a.uponDecisionInputs.ccOutputReceived(*ccOut.(*bool)))
		}
	}
	return msgs
}

// >     – wait until bin_values_r != {}, then
// >         ∗ multicast AUX_r(w) where w ∈ bin_values_r
// >         ∗ wait until at least (N − f) AUX_r messages have been
// >           received, such that the set of values carried by these
// >           messages, vals are a subset of bin_values_r (note that
// >           bin_values_r may continue to change as BVAL_r messages
// >           are received, thus this condition may be triggered upon
// >           arrival of either an AUX_r or a BVAL_r message)
func (a *ABA) uponBinValuesUpdated(binValues []bool) []gpa.MessageOut {
	return a.varAuxVals.binValuesUpdated(binValues)
}

// >         ∗ wait until at least (N − f) AUX_r messages have been
// >           received, such that the set of values carried by these
// >           messages, vals are a subset of bin_values_r (note that
// >           bin_values_r may continue to change as BVAL_r messages
// >           are received, thus this condition may be triggered upon
// >           arrival of either an AUX_r or a BVAL_r message)
func (a *ABA) uponAuxValsReady(auxVals []bool) []gpa.MessageOut {
	return a.uponDecisionInputs.auxValsReady(auxVals)
}

// >         ∗ if vals = {b}, then
// >             · est_r+1 := b
// >             · if (b = s%2) then output b
// >         ∗ else est_r+1 := s%2
func (a *ABA) uponDecisionInputsReceived(cc bool, auxVals []bool) []gpa.MessageOut {
	if len(auxVals) == 1 {
		nextEst := auxVals[0]
		if nextEst == cc {
			if a.output == nil {
				a.output = &Output{Value: nextEst, Terminated: a.varDone.isDone()}
			}
			return slices.Concat(
				a.varDone.outputProduced(),
				a.startRound(a.round+1, nextEst),
			)
		}
		return a.startRound(a.round+1, nextEst)
	}
	return a.startRound(a.round+1, cc)
}

// Here we get notification from `varDone` on the termination.
func (a *ABA) uponTerminationCondition() {
	if a.output != nil {
		a.output.Terminated = true
	}
}

// Output implements the gpa.GPA interface.
func (a *ABA) Output() gpa.Output {
	if a.output == nil {
		return nil // Untyped nil
	}
	return a.output
}

// StatusString implements the gpa.GPA interface.
func (a *ABA) StatusString() string {
	return fmt.Sprintf(
		"{ABA:Mostefaoui, R=%v, %v, %v, %v, %v, out=%+v}",
		a.round,
		a.varBinVals.statusString(),
		a.varAuxVals.statusString(),
		a.uponDecisionInputs.statusString(),
		a.varDone.statusString(),
		a.output,
	)
}
