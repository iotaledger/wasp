// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/state"
)

type inputStateMgrDecidedVirtualState struct {
	chainState state.State
}

func NewInputStateMgrDecidedVirtualState(
	chainState state.State,
) gpa.Input {
	return &inputStateMgrDecidedVirtualState{chainState: chainState}
}
