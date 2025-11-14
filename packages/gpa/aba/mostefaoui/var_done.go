// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mostefaoui

import (
	"fmt"
	"slices"

	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/v2/packages/gpa"
)

// We can terminate the algorithm if:
//   - Either that's the second time we deciding.
//   - That's the first time we deciding, but we have F+1
//     DONE messages that others decided before.
//
// The DONE messages can be late. To handle the second case,
// after receiving a DONE message we have to check if our last
// decision is after the F+1 DONE messages.
type varDone struct {
	nodeIDs []gpa.NodeID
	me      gpa.NodeID
	f       int
	round   int
	recv    map[gpa.NodeID]int // All the received DONE messages and last our decision.
	doneCB  func()
	done    bool
	log     log.Logger
}

func newVarDone(nodeIDs []gpa.NodeID, me gpa.NodeID, f int, doneCB func(), log log.Logger) *varDone {
	return &varDone{
		nodeIDs: nodeIDs,
		me:      me,
		f:       f,
		round:   -1,
		recv:    map[gpa.NodeID]int{},
		doneCB:  doneCB,
		done:    false,
		log:     log,
	}
}

func (v *varDone) startRound(round int) {
	v.round = round
}

func (v *varDone) setDone() {
	if !v.done {
		v.done = true
		v.doneCB()
	}
}

func (v *varDone) outputProduced() []gpa.PayloadOut {
	if firstDoneRound, ok := v.recv[v.me]; ok && firstDoneRound < v.round {
		// We have decided for the second time. That's enough.
		v.setDone()
		return nil
	}

	v.recv[v.me] = v.round
	return slices.Concat(
		multicastMsgDone(v.nodeIDs, v.me, v.round),
		v.tryComplete(),
	)
}

func (v *varDone) msgDoneReceived(msg gpa.PayloadIn[MsgDone]) []gpa.PayloadOut {
	if _, ok := v.recv[msg.Sender]; ok {
		return nil // Duplicate
	}
	v.recv[msg.Sender] = msg.Payload.round
	return v.tryComplete()
}

func (v *varDone) isDone() bool {
	return v.done
}

// If others (more than F) have decided in previous epochs, then we are
// among the others, who decided in a subsequent round, therefore we don't
// need to wait for more epochs to close the process.
func (v *varDone) tryComplete() []gpa.PayloadOut {
	if v.done || len(v.recv) <= v.f {
		return nil
	}
	outDecidedRound, ok := v.recv[v.me]
	if !ok {
		// We have not decided yet, can't close the process.
		return nil
	}
	count := 0
	for _, r := range v.recv {
		if r < outDecidedRound {
			count++
		}
	}
	if count > v.f {
		v.setDone()
	}
	return nil
}

func (v *varDone) statusString() string {
	return fmt.Sprintf("|done|=%v/%v=%v", len(v.recv), len(v.nodeIDs), v.done)
}
