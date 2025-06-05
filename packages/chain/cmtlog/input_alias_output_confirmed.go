// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtlog

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type inputAnchorConfirmed struct {
	anchor *isc.StateAnchor
}

func NewInputAnchorConfirmed(anchor *isc.StateAnchor) gpa.Input {
	return &inputAnchorConfirmed{
		anchor: anchor,
	}
}

func (inp *inputAnchorConfirmed) String() string {
	return fmt.Sprintf("{cmtLog.inputAnchorConfirmed, %v}", inp.anchor)
}
