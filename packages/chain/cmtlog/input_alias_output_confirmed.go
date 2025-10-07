// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtlog

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/isc"
)

type inputAnchorConfirmed struct {
	anchor *isc.StateAnchor
}

func NewInputAnchorConfirmed(anchor *isc.StateAnchor) *inputAnchorConfirmed {
	return &inputAnchorConfirmed{
		anchor: anchor,
	}
}

func (inp *inputAnchorConfirmed) String() string {
	return fmt.Sprintf("{cmtLog.inputAnchorConfirmed, %v}", inp.anchor)
}
