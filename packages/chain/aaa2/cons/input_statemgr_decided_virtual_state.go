// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/state"
)

type inputStateMgrDecidedVirtualState struct {
	stateBaseline      coreutil.StateBaseline
	virtualStateAccess state.VirtualStateAccess
}

func NewInputStateMgrDecidedVirtualState(
	stateBaseline coreutil.StateBaseline,
	virtualStateAccess state.VirtualStateAccess,
) gpa.Input {
	return &inputStateMgrDecidedVirtualState{stateBaseline, virtualStateAccess}
}
