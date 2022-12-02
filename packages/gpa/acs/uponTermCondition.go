// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package acs

import "github.com/iotaledger/wasp/packages/gpa"

// Here we track the termination condition.
type uponTermCondition struct {
	n      int
	term   map[gpa.NodeID]bool
	termCB func() gpa.OutMessages
	done   bool
}

func newUponTermCondition(n int, termCB func() gpa.OutMessages) *uponTermCondition {
	return &uponTermCondition{
		n:      n,
		term:   map[gpa.NodeID]bool{},
		termCB: termCB,
		done:   false,
	}
}

func (u *uponTermCondition) abaTerminated(nodeID gpa.NodeID) gpa.OutMessages {
	if u.done {
		return nil
	}
	if ok := u.term[nodeID]; ok {
		return nil
	}
	u.term[nodeID] = true
	if len(u.term) == u.n {
		u.done = true
		return u.termCB()
	}
	return nil
}

func (u *uponTermCondition) canTerminate() bool {
	return u.done
}
