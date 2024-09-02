// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainmanager

import (
	"fmt"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
)

type inputAliasOutputConfirmed struct {
	stateController *cryptolib.Address
	anchor          *iscmove.Anchor
}

func NewInputAliasOutputConfirmed(stateController *cryptolib.Address, anchor *iscmove.Anchor) gpa.Input {
	return &inputAliasOutputConfirmed{
		stateController: stateController,
		anchor:          anchor,
	}
}

func (inp *inputAliasOutputConfirmed) String() string {
	return fmt.Sprintf("{chainMgr.inputAliasOutputConfirmed, %v}", inp.anchor.ID.ShortString())
}
