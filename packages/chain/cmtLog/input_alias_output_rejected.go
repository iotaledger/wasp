// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtLog

import (
	"fmt"

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

func (inp *inputAliasOutputRejected) String() string {
	return fmt.Sprintf("{cmtLog.inputAliasOutputRejected, %v}", inp.aliasOutput)
}
