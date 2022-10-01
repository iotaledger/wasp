// Copyright 2020 IOTA Stiftung
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

	"golang.org/x/xerrors"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/gpa"
)

// This structure is provided as an output of the algorithm.
// If the value is undecided, untyped nil is returned.
// The Terminate field indicates, if this algorithm can be
// dropped (no other peers need any messages from this node).
type Output struct {
	Value      bool
	Terminated bool
}

// Public API for this protocol.
type ABA interface {
	AsGPA() gpa.GPA
}

const (
	subsystemCC byte = iota
)

type abaImpl struct {
	nodeIDs            []gpa.NodeID            // Nodes in the consensus.
	nodeIdx            map[gpa.NodeID]bool     // For a fast check, if peer is known.
	round              int                     // The current round.
	varBinVals         *varBinVals             // The `binValues` variable (based on BVAL msgs).
	varAuxVals         *varAuxVals             // The `vals` variable (based on AUX msgs).
	varDone            *varDone                // Termination condition.
	uponDecisionInputs *uponDecisionInputs     // Decision condition.
	ccInsts            []gpa.GPA               // Common coin instances for all the rounds.
	ccCreateFun        func(round int) gpa.GPA // Function to create CC instances.
	output             *Output                 // The current output of the algorithm.
	postponedMsgs      []*msgVote              // Buffer for future round messages.
	msgWrapper         *gpa.MsgWrapper         // Helper to wrap messages for sub-components.
	asGPA              gpa.GPA                 // This object, but with required wrappers.
	log                *logger.Logger          // A logger.
}

var (
	_ gpa.GPA = &abaImpl{}
	_ ABA     = &abaImpl{}
)

// Creates a single node for a consensus.
//
// Here `ccCreateFun` is used as a factory function to create Common Coin instances for each round.
// This way this implementation is made independent of particular CC instance. The created CC
// is expected to take `nil` inputs and produce `*bool` outputs.
func New(nodeIDs []gpa.NodeID, me gpa.NodeID, f int, ccCreateFun func(round int) gpa.GPA, log *logger.Logger) ABA {
	nodeIdx := map[gpa.NodeID]bool{}
	for _, n := range nodeIDs {
		nodeIdx[n] = true
	}
	a := &abaImpl{
		nodeIDs:       nodeIDs,
		nodeIdx:       nodeIdx,
		round:         -1,
		ccInsts:       []gpa.GPA{},
		ccCreateFun:   ccCreateFun,
		output:        nil,
		postponedMsgs: []*msgVote{},
		log:           log,
	}
	a.varBinVals = newBinVals(nodeIDs, f, a.uponBinValuesUpdated)
	a.varAuxVals = newAuxVals(nodeIDs, f, a.uponAuxValsReady)
	a.varDone = newVarDone(nodeIDs, me, f, a.uponTerminationCondition, log)
	a.uponDecisionInputs = newUponDecisionInputs(a.uponDecisionInputsReceived)
	a.msgWrapper = gpa.NewMsgWrapper(msgTypeWrapped, a.selectSubsystem)
	a.asGPA = gpa.NewOwnHandler(me, a)
	return a
}

// Helper for routing messages to sub-protocols (i.e. CC instances).
func (a *abaImpl) selectSubsystem(subsystem byte, index int) (gpa.GPA, error) {
	if subsystem == subsystemCC {
		if index > a.round+10 {
			return nil, xerrors.Errorf("cc round=%v to far in future, our round=%v", index, a.round)
		}
		return a.ccInst(index), nil
	}
	return nil, xerrors.Errorf("unexpected subsystem=%v, index=%v", subsystem, index)
}

// Creates and returns a CC instance for a particular round.
// CC instances are not cleaned up, as the algorithm is supposed to terminate in few rounds.
func (a *abaImpl) ccInst(round int) gpa.GPA {
	if round >= len(a.ccInsts) {
		add := make([]gpa.GPA, round-len(a.ccInsts)+1)
		a.ccInsts = append(a.ccInsts, add...)
	}
	if a.ccInsts[round] == nil {
		a.ccInsts[round] = a.ccCreateFun(round)
	}
	return a.ccInsts[round]
}

// Implements the ABA interface.
func (a *abaImpl) AsGPA() gpa.GPA {
	return a.asGPA
}

// Implements the gpa.GPA interface.
//
// > • upon receiving input b_input, set est_0 := b_input and proceed as
// >   follows in consecutive epochs, with increasing labels r:
func (a *abaImpl) Input(input gpa.Input) gpa.OutMessages {
	return a.startRound(0, input.(bool))
}

// Advances the algorithm to the next round.
//
// >     – multicast BVAL_r(est_r)
// >     – bin_values_r := {}
func (a *abaImpl) startRound(round int, est bool) gpa.OutMessages {
	if a.output != nil && a.output.Terminated {
		// Don't start the next round if the algorithm is already terminated.
		return nil
	}
	msgs := gpa.NoMessages()
	a.round = round
	a.varAuxVals.startRound(a.round)
	a.varDone.startRound(round)
	a.uponDecisionInputs.startRound()
	msgs.AddAll(a.varBinVals.startRound(a.round, est))
	//
	// Start the CC.
	ccInst := a.ccInst(a.round)
	msgs.AddAll(a.msgWrapper.WrapMessages(subsystemCC, round, ccInst.Input(nil)))
	if out := ccInst.Output(); out != nil {
		msgs.AddAll(a.uponDecisionInputs.ccOutputReceived(*out.(*bool)))
	}
	//
	// Resend postponed messages, if any.
	if len(a.postponedMsgs) > 0 {
		oldPostponedMsgs := a.postponedMsgs
		a.postponedMsgs = []*msgVote{}
		for _, m := range oldPostponedMsgs {
			msgs.AddAll(a.Message(m))
		}
	}
	return msgs
}

// Implements the gpa.GPA interface.
// Here we only route the messages to appropriate objects.
func (a *abaImpl) Message(msg gpa.Message) gpa.OutMessages {
	switch msgT := msg.(type) {
	case *msgVote: // The BVAL and AUX messages.
		if _, ok := a.nodeIdx[msgT.sender]; !ok {
			return nil // Unknown sender.
		}
		if msgT.round < a.round {
			return nil // Outdated message.
		}
		if msgT.round > a.round {
			a.postponedMsgs = append(a.postponedMsgs, msgT)
			return nil // Will be processed later.
		}
		switch msgT.voteType {
		case BVAL:
			return a.varBinVals.msgVoteBVALReceived(msgT)
		case AUX:
			return a.varAuxVals.msgVoteAUXReceived(msgT)
		}
		a.log.Warnf("unexpected msgVote message: %+v", msgT)
		return nil
	case *msgDone: // The DONE messages for the termination.
		if _, ok := a.nodeIdx[msgT.sender]; !ok {
			return nil // Unknown sender.
		}
		return a.varDone.msgDoneReceived(msgT)
	case *gpa.WrappingMsg: // The CC messages.
		sub, err := a.selectSubsystem(msgT.Subsystem(), msgT.Index())
		if err != nil {
			a.log.Warnf("cannot select subsystem: %v", err)
			return nil
		}
		msgs := gpa.NoMessages()
		msgs.AddAll(a.msgWrapper.WrapMessages(msgT.Subsystem(), msgT.Index(), sub.Message(msgT.Wrapped())))
		if msgT.Subsystem() == subsystemCC && msgT.Index() == a.round && !a.uponDecisionInputs.haveCC() {
			ccInst := a.ccInst(msgT.Index())
			ccOut := ccInst.Output()
			if ccOut != nil {
				msgs.AddAll(a.uponDecisionInputs.ccOutputReceived(*ccOut.(*bool)))
			}
		}
		return msgs
	}
	a.log.Warnf("unexpected message of type %T: %+v", msg, msg)
	return nil
}

// >     – wait until bin_values_r != {}, then
// >         ∗ multicast AUX_r(w) where w ∈ bin_values_r
// >         ∗ wait until at least (N − f) AUX_r messages have been
// >           received, such that the set of values carried by these
// >           messages, vals are a subset of bin_values_r (note that
// >           bin_values_r may continue to change as BVAL_r messages
// >           are received, thus this condition may be triggered upon
// >           arrival of either an AUX_r or a BVAL_r message)
func (a *abaImpl) uponBinValuesUpdated(binValues []bool) gpa.OutMessages {
	return a.varAuxVals.binValuesUpdated(binValues)
}

// >         ∗ wait until at least (N − f) AUX_r messages have been
// >           received, such that the set of values carried by these
// >           messages, vals are a subset of bin_values_r (note that
// >           bin_values_r may continue to change as BVAL_r messages
// >           are received, thus this condition may be triggered upon
// >           arrival of either an AUX_r or a BVAL_r message)
func (a *abaImpl) uponAuxValsReady(auxVals []bool) gpa.OutMessages {
	return a.uponDecisionInputs.auxValsReady(auxVals)
}

// >         ∗ if vals = {b}, then
// >             · est_r+1 := b
// >             · if (b = s%2) then output b
// >         ∗ else est_r+1 := s%2
func (a *abaImpl) uponDecisionInputsReceived(cc bool, auxVals []bool) gpa.OutMessages {
	if len(auxVals) == 1 {
		nextEst := auxVals[0]
		if nextEst == cc {
			if a.output == nil {
				a.output = &Output{Value: nextEst, Terminated: a.varDone.isDone()}
			}
			msgs := gpa.NoMessages()
			msgs.AddAll(a.varDone.outputProduced())
			return msgs.AddAll(a.startRound(a.round+1, nextEst))
		}
		return a.startRound(a.round+1, nextEst)
	}
	return a.startRound(a.round+1, cc)
}

// Here we get notification from `varDone` on the termination.
func (a *abaImpl) uponTerminationCondition() gpa.OutMessages {
	if a.output != nil {
		a.output.Terminated = true
	}
	return nil
}

// Implements the gpa.GPA interface.
func (a *abaImpl) Output() gpa.Output {
	if a.output == nil {
		return nil // Untyped nil
	}
	return a.output
}

// Implements the gpa.GPA interface.
func (a *abaImpl) StatusString() string {
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
