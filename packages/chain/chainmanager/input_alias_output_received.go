// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainmanager

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type inputAliasOutputConfirmed struct {
	stateController *cryptolib.Address
	anchor          *isc.StateAnchor
}

func NewInputAliasOutputConfirmed(stateController *cryptolib.Address, anchor *isc.StateAnchor) gpa.Input {
	return &inputAliasOutputConfirmed{
		stateController: stateController,
		anchor:          anchor,
	}
}

func (inp *inputAliasOutputConfirmed) String() string {
	return fmt.Sprintf("{chainMgr.inputAliasOutputConfirmed, %v}", inp.anchor.GetObjectID().ShortString())
}
