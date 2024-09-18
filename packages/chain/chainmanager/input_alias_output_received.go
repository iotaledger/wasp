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
	anchor          *iscmove.AnchorWithRef
}

func NewInputAliasOutputConfirmed(stateController *cryptolib.Address, anchor *iscmove.AnchorWithRef) gpa.Input {
	return &inputAliasOutputConfirmed{
		stateController: stateController,
		anchor:          anchor,
	}
}

func (inp *inputAliasOutputConfirmed) String() string {
	return fmt.Sprintf("{chainMgr.inputAliasOutputConfirmed, %v}", inp.anchor.Object.ID.ShortString())
}
