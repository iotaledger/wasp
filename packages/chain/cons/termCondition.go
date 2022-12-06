// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

type termCondition struct {
	acsTerminated  bool
	outputProduced bool
	done           bool
	doneCB         func()
}

func newTermCondition(doneCB func()) *termCondition {
	return &termCondition{doneCB: doneCB}
}

func (tc *termCondition) haveAcsTerminated() {
	if tc.acsTerminated {
		return
	}
	tc.acsTerminated = true
	tc.tryComplete()
}

func (tc *termCondition) haveOutputProduced() {
	if tc.outputProduced {
		return
	}
	tc.outputProduced = true
	tc.tryComplete()
}

func (tc *termCondition) tryComplete() {
	if tc.done || !tc.acsTerminated || !tc.outputProduced {
		return
	}
	tc.done = true
	tc.doneCB()
}
