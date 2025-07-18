// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainmanager

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
)

type inputAnchorConfirmed struct {
	stateController *cryptolib.Address
	anchor          *isc.StateAnchor
}

func NewInputAnchorConfirmed(stateController *cryptolib.Address, anchor *isc.StateAnchor) gpa.Input {
	return &inputAnchorConfirmed{
		stateController: stateController,
		anchor:          anchor,
	}
}

func (inp *inputAnchorConfirmed) String() string {
	return fmt.Sprintf("{chainMgr.inputAnchorConfirmed, %v}", inp.anchor.GetObjectID().ShortString())
}
