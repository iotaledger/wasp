// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mostefaoui

import (
	"fmt"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/gpa"
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
	doneCB  func() gpa.OutMessages
	done    bool
	log     *logger.Logger
}

func newVarDone(nodeIDs []gpa.NodeID, me gpa.NodeID, f int, doneCB func() gpa.OutMessages, log *logger.Logger) *varDone {
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

func (v *varDone) outputProduced() gpa.OutMessages {
	if _, ok := v.recv[v.me]; ok {
		// We have decided for the second time. That's enough.
		if !v.done {
			v.done = true
			return v.doneCB()
		}
		return nil
	}

	v.recv[v.me] = v.round
	msgs := gpa.NoMessages()
	msgs.AddAll(multicastMsgDone(v.nodeIDs, v.me, v.round))
	msgs.AddAll(v.tryComplete())
	return msgs
}

func (v *varDone) msgDoneReceived(msg *msgDone) gpa.OutMessages {
	if _, ok := v.recv[msg.sender]; ok {
		return nil // Duplicate
	}
	v.recv[msg.sender] = msg.round
	return v.tryComplete()
}

func (v *varDone) isDone() bool {
	return v.done
}

// If others (more than F) have decided in previous epochs, then we are
// among the others, who decided in a subsequent round, therefore we don't
// need to wait for more epochs to close the process.
func (v *varDone) tryComplete() gpa.OutMessages {
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
		v.done = true
		return v.doneCB()
	}
	return nil
}

func (v *varDone) statusString() string {
	return fmt.Sprintf("|done|=%v/%v=%v", len(v.nodeIDs), len(v.recv), v.done)
}
