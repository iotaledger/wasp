// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package mostefaoui

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
)

// This object implements a synchronization before the decision.
// We have to wait for `vals` (auxVals) and `cc` before deciding.
// We notify the upped algorithm via the callback exactly once.
//
// >         ∗ wait until at least (N − f) AUX_r messages have been
// >           received, such that the set of values carried by these
// >           messages, vals are a subset of bin_values_r (note that
// >           bin_values_r may continue to change as BVAL_r messages
// >           are received, thus this condition may be triggered upon
// >           arrival of either an AUX_r or a BVAL_r message)
// >         ∗ s ← Coin_r.GetCoin()
type uponDecisionInputs struct {
	ccReceived bool
	ccValue    bool
	auxVals    []bool
	done       bool
	doneCB     func(cc bool, auxVals []bool) gpa.OutMessages
}

func newUponDecisionInputs(doneCB func(cc bool, auxVals []bool) gpa.OutMessages) *uponDecisionInputs {
	u := &uponDecisionInputs{doneCB: doneCB}
	u.startRound()
	return u
}

func (u *uponDecisionInputs) startRound() {
	u.ccReceived = false
	u.ccValue = false
	u.auxVals = nil
	u.done = false
}

func (u *uponDecisionInputs) ccOutputReceived(cc bool) gpa.OutMessages {
	if u.ccReceived {
		return nil
	}
	u.ccValue = cc
	u.ccReceived = true
	return u.tryOutput()
}

func (u *uponDecisionInputs) auxValsReady(auxVals []bool) gpa.OutMessages {
	if u.auxVals != nil {
		return nil
	}
	u.auxVals = auxVals
	return u.tryOutput()
}

func (u *uponDecisionInputs) tryOutput() gpa.OutMessages {
	if u.done || !u.ccReceived || u.auxVals == nil {
		return nil
	}
	u.done = true
	return u.doneCB(u.ccValue, u.auxVals)
}

func (u *uponDecisionInputs) haveCC() bool {
	return u.ccReceived
}

func (u *uponDecisionInputs) statusString() string {
	if u.ccReceived {
		return fmt.Sprintf("CC=%v, auxVals=%v", u.ccValue, u.auxVals)
	}
	return fmt.Sprintf("CC=nil, auxVals=%v", u.auxVals)
}
