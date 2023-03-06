// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package distSync

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type inputRequestNeeded struct {
	requestRef *isc.RequestRef
	needed     bool // false to cancel.
}

func NewInputRequestNeeded(requestRef *isc.RequestRef, needed bool) gpa.Input {
	return &inputRequestNeeded{requestRef: requestRef, needed: needed}
}

func (inp *inputRequestNeeded) String() string {
	return fmt.Sprintf("{mp.ds.inputRequestNeeded, requestRef=%v, needed=%v}", inp.requestRef, inp.needed)
}
