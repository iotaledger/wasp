// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// Here we implement the Asynchronous Common Subset algorithm
// from the HBBFT paper:
//
// > Miller, A., Xia, Y., Croman, K., Shi, E., and Song, D. (2016). The Honey Badger of
// > BFT Protocols. In Proceedings of the 2016 ACM SIGSAC Conference on Computer
// > and Communications Security, CCS ’16, page 31–42, New York, NY, USA.
// > Association for Computing Machinery.
//
// The HBBFT paper presents the algorithm as follows:
//
// > Let {RBC_i}_N refer to N instances of the reliable broadcast protocol,
// > where P_i is the sender of RBC_i. Let {BA_i}_N refer to N instances
// > of the binary byzantine agreement protocol.
// >   • upon receiving input v_i, input v_i to RBC_i
// >   • upon delivery of v_j from RBC_j, if input has not yet been
// >     provided to BA_j, then provide input 1 to BA_j.
// >   • upon delivery of value 1 from at least N − f instances of BA,
// >     provide input 0 to each instance of BA that has not yet been
// >     provided input.
// >   • once all instances of BA have completed, let C ⊂ [1..N] be the
// >     indexes of each BA that delivered 1. Wait for the output v_j for
// >     each RBC_j such that j ∈ C. Finally output ∪_{j∈C} v_j.
//
// TODO: Erasure coding in RBC.
package acs

import (
	"fmt"
	"math"

	"golang.org/x/xerrors"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/gpa/aba/mostefaoui"
	"github.com/iotaledger/wasp/packages/gpa/rbc/bracha"
)

type ACS interface {
	AsGPA() gpa.GPA
}

type Output struct {
	Values     map[gpa.NodeID][]byte
	Terminated bool
}

const (
	subsystemRBC byte = iota
	subsystemABA
)

type acsImpl struct {
	nodeIDs    []gpa.NodeID           // Nodes in the consensus.
	nodeIdx    map[gpa.NodeID]int     // For a fast check, if peer is known.
	me         gpa.NodeID             // Out name.
	n          int                    // Number of nodes in the cluster.
	f          int                    // Max number of tolerated faulty nodes.
	rbcInsts   map[gpa.NodeID]gpa.GPA // RBC instances.
	rbcInput   bool                   // Have we provided our input?
	rbcOutputs map[gpa.NodeID][]byte  // Outputs received from the RBC.
	abaInsts   map[gpa.NodeID]gpa.GPA // ABA Instances.
	abaInputs  map[gpa.NodeID]bool    // Inputs already provided to ABAs.
	abaOutputs map[gpa.NodeID]bool    // Outputs already received from ABAs.
	output     *Output                // Output we produced.
	termCond   *uponTermCondition     // Tracks the termination condition.
	msgWrapper *gpa.MsgWrapper        // Helper to wrap messages for sub-components.
	asGPA      gpa.GPA                // This object with required wrappers.
	log        *logger.Logger         // A logger.
}

var (
	_ gpa.GPA = &acsImpl{}
	_ ACS     = &acsImpl{}
)

// > Let {RBC_i}_N refer to N instances of the reliable broadcast protocol,
// > where P_i is the sender of RBC_i. Let {BA_i}_N refer to N instances
// > of the binary byzantine agreement protocol.
func New(nodeIDs []gpa.NodeID, me gpa.NodeID, f int, ccCreateFun func(node gpa.NodeID, round int) gpa.GPA, log *logger.Logger) ACS {
	nodeIdx := map[gpa.NodeID]int{}
	rbcInsts := map[gpa.NodeID]gpa.GPA{}
	abaInsts := map[gpa.NodeID]gpa.GPA{}
	for i, nid := range nodeIDs {
		nidCopy := nid
		ccCreateFunForNode := func(round int) gpa.GPA {
			return ccCreateFun(nidCopy, round)
		}
		nodeIdx[nid] = i
		rbcInsts[nid] = bracha.New(nodeIDs, f, me, nid, math.MaxInt, func(b []byte) bool { return true }) // TODO: MaxInt.
		abaInsts[nid] = mostefaoui.New(nodeIDs, me, f, ccCreateFunForNode, log).AsGPA()
	}

	n := len(nodeIDs)
	a := &acsImpl{
		nodeIDs:    nodeIDs,
		nodeIdx:    nodeIdx,
		me:         me,
		n:          n,
		f:          f,
		rbcInsts:   rbcInsts,
		rbcInput:   false,
		rbcOutputs: map[gpa.NodeID][]byte{},
		abaInsts:   abaInsts,
		abaInputs:  map[gpa.NodeID]bool{},
		abaOutputs: map[gpa.NodeID]bool{},
		output:     nil,
		log:        log,
	}
	a.termCond = newUponTermCondition(n, a.uponTermCondition)
	a.msgWrapper = gpa.NewMsgWrapper(msgTypeWrapped, a.selectSubsystem)
	a.asGPA = gpa.NewOwnHandler(me, a)
	return a
}

// Helper for routing messages to sub-protocols (i.e. RBC and ABA instances).
func (a *acsImpl) selectSubsystem(subsystem byte, index int) (gpa.GPA, error) {
	if index < 0 || index >= a.n {
		return nil, xerrors.Errorf("unexpected index=%v for subsystem", index)
	}
	nid := a.nodeIDs[index]
	switch subsystem {
	case subsystemRBC:
		return a.rbcInsts[nid], nil
	case subsystemABA:
		return a.abaInsts[nid], nil
	}
	return nil, xerrors.Errorf("unexpected subsystem=%v, index=%v", subsystem, index)
}

func (a *acsImpl) AsGPA() gpa.GPA {
	return a.asGPA
}

// >   • upon receiving input v_i, input v_i to RBC_i
func (a *acsImpl) Input(input gpa.Input) gpa.OutMessages {
	if _, ok := input.([]byte); !ok {
		panic("input has to be []byte")
	}
	if a.rbcInput {
		return nil // Duplicate input.
	}
	a.rbcInput = true
	msgs := gpa.NoMessages()
	sub, subMsgs, err := a.msgWrapper.DelegateInput(subsystemRBC, a.nodeIdx[a.me], input)
	if err != nil {
		panic(xerrors.Errorf("cannot provide input to RBC: %w", err))
	}
	msgs.AddAll(subMsgs)
	msgs.AddAll(a.tryHandleRBCOutput(a.me, sub))
	return msgs
}

func (a *acsImpl) Message(msg gpa.Message) gpa.OutMessages {
	msgT, ok := msg.(*gpa.WrappingMsg)
	if !ok {
		a.log.Warn("unexpected message of type %T: %+v", msg, msg)
		return nil
	}
	msgs := gpa.NoMessages()
	sub, subMsgs, err := a.msgWrapper.DelegateMessage(*msgT)
	if err != nil {
		a.log.Warn("cannot delegate a message: %v", err)
		return nil
	}
	msgs.AddAll(subMsgs)
	switch msgT.Subsystem() {
	case subsystemRBC:
		msgs.AddAll(a.tryHandleRBCOutput(a.nodeIDs[msgT.Index()], sub))
		return msgs
	case subsystemABA:
		msgs.AddAll(a.tryHandleABAOutput(a.nodeIDs[msgT.Index()], sub))
		return msgs
	}
	panic(xerrors.Errorf("unexpected subsystem: %v", msgT.Subsystem()))
}

// >   • upon delivery of v_j from RBC_j, if input has not yet been
// >     provided to BA_j, then provide input 1 to BA_j.
func (a *acsImpl) tryHandleRBCOutput(nodeID gpa.NodeID, rbcInst gpa.GPA) gpa.OutMessages {
	out := rbcInst.Output()
	if out == nil {
		return nil // Output not ready yet.
	}
	if _, ok := a.rbcOutputs[nodeID]; ok {
		return nil // Already handled.
	}
	a.rbcOutputs[nodeID] = out.([]byte)
	a.tryOutput()

	if _, ok := a.abaInputs[nodeID]; ok {
		return nil // We already provided an input to the ABA.
	}
	a.abaInputs[nodeID] = true
	msgs := gpa.NoMessages()
	sub, subMsgs, err := a.msgWrapper.DelegateInput(subsystemABA, a.nodeIdx[nodeID], true)
	if err != nil {
		panic(xerrors.Errorf("cannot provide input to ABA: %w", err))
	}
	msgs.AddAll(subMsgs)
	msgs.AddAll(a.tryHandleABAOutput(nodeID, sub))
	return msgs
}

// >   • upon delivery of value 1 from at least N − f instances of BA,
// >     provide input 0 to each instance of BA that has not yet been
// >     provided input.
func (a *acsImpl) tryHandleABAOutput(nodeID gpa.NodeID, abaInst gpa.GPA) gpa.OutMessages {
	out := abaInst.Output()
	if out == nil {
		return nil // Output not ready yet.
	}
	abaOut := out.(*mostefaoui.Output)
	msgs := gpa.NoMessages()
	if abaOut.Terminated {
		msgs.AddAll(a.termCond.abaTerminated(nodeID))
	}

	if _, ok := a.abaOutputs[nodeID]; ok {
		return msgs // Already handled.
	}
	a.abaOutputs[nodeID] = abaOut.Value
	a.tryOutput()

	if len(a.abaOutputs) < a.n-a.f || len(a.abaInputs) == a.n {
		return msgs
	}
	count := 0
	for _, abaOut := range a.abaOutputs {
		if abaOut {
			count++
		}
	}
	if count >= a.n-a.f {
		for _, nid := range a.nodeIDs {
			if ok := a.abaInputs[nid]; ok {
				continue // Input was already provided.
			}
			a.abaInputs[nid] = false
			sub, subMsgs, err := a.msgWrapper.DelegateInput(subsystemABA, a.nodeIdx[nid], false)
			if err != nil {
				panic(xerrors.Errorf("cannot provide input to ABA: %w", err))
			}
			msgs.AddAll(subMsgs)
			msgs.AddAll(a.tryHandleABAOutput(nid, sub))
		}
	}
	return msgs
}

// >   • once all instances of BA have completed, let C ⊂ [1..N] be the
// >     indexes of each BA that delivered 1. Wait for the output v_j for
// >     each RBC_j such that j ∈ C. Finally output ∪_{j∈C} v_j.
func (a *acsImpl) tryOutput() {
	if a.output != nil {
		return // Output already provided.
	}
	if len(a.abaOutputs) < a.n {
		return // Not all ABAs have provided an output.
	}
	values := map[gpa.NodeID][]byte{}
	for nid, abaOut := range a.abaOutputs {
		if abaOut {
			if rbcOut, ok := a.rbcOutputs[nid]; ok {
				values[nid] = rbcOut
				continue
			}
			return // Some RBC output are still missing.
		}
	}
	a.output = &Output{
		Values:     values,
		Terminated: a.termCond.canTerminate(),
	}
}

func (a *acsImpl) uponTermCondition() gpa.OutMessages {
	if a.output != nil {
		a.output.Terminated = true
	}
	return nil
}

func (a *acsImpl) Output() gpa.Output {
	if a.output == nil {
		return nil // Untyped nil.
	}
	return a.output
}

func (a *acsImpl) StatusString() string {
	if a.output != nil {
		return fmt.Sprintf(
			"{ACS, |outVals|=%v, outTerm=%+v, n=%v, f=%v, |rbcOut|=%v, |abaOut|=%v}",
			len(a.output.Values), a.output.Terminated, a.n, a.f, len(a.rbcOutputs), len(a.abaOutputs),
		)
	}
	return fmt.Sprintf(
		"{ACS, out=nil, n=%v, f=%v, |rbcOut|=%v, |abaOut|=%v}",
		a.n, a.f, len(a.rbcOutputs), len(a.abaOutputs),
	)
}
