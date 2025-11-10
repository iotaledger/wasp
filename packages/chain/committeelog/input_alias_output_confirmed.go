// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package committeelog

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/isc"
)

type InputAnchorConfirmed struct {
	anchor *isc.StateAnchor
}

func NewInputAnchorConfirmed(anchor *isc.StateAnchor) *InputAnchorConfirmed {
	return &InputAnchorConfirmed{
		anchor: anchor,
	}
}

func (inp *InputAnchorConfirmed) String() string {
	return fmt.Sprintf("{committeeLog.inputAnchorConfirmed, %v}", inp.anchor)
}
