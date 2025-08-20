// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/state"
)

type inputStateMgrBlockSaved struct {
	block state.Block
}

func NewInputStateMgrBlockSaved(block state.Block) gpa.Input {
	return &inputStateMgrBlockSaved{block: block}
}

func (inp *inputStateMgrBlockSaved) String() string {
	return fmt.Sprintf("{cons.inputStateMgrBlockSaved, stateIndex=%v, l1Commitment=%v}", inp.block.StateIndex(), inp.block.L1Commitment())
}
