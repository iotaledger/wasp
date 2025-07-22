// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package distsync

import (
	"context"
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
)

type inputRequestNeeded struct {
	ctx        context.Context
	requestRef *isc.RequestRef
}

func NewInputRequestNeeded(ctx context.Context, requestRef *isc.RequestRef) gpa.Input {
	return &inputRequestNeeded{ctx: ctx, requestRef: requestRef}
}

func (inp *inputRequestNeeded) String() string {
	return fmt.Sprintf("{mp.ds.inputRequestNeeded, requestRef=%v}", inp.requestRef)
}
