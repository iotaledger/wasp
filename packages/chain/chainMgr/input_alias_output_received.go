// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chainMgr

import (
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type inputAliasOutputConfirmed struct {
	aliasOutput *isc.AliasOutputWithID
}

func NewInputAliasOutputConfirmed(aliasOutput *isc.AliasOutputWithID) gpa.Input {
	return &inputAliasOutputConfirmed{
		aliasOutput: aliasOutput,
	}
}
