// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtLog

import (
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

type inputAliasOutputRejected struct {
	aliasOutput *isc.AliasOutputWithID
}

func NewInputAliasOutputRejected(aliasOutput *isc.AliasOutputWithID) gpa.Input {
	return &inputAliasOutputRejected{
		aliasOutput: aliasOutput,
	}
}
