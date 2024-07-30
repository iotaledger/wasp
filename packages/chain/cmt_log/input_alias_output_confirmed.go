// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmt_log

import (
	"fmt"

	"github.com/iotaledger/wasp/clients/iscmove/isctypes"
	"github.com/iotaledger/wasp/packages/gpa"
)

type inputAliasOutputConfirmed struct {
	aliasOutput *isctypes.Anchor
}

func NewInputAliasOutputConfirmed(aliasOutput *isctypes.Anchor) gpa.Input {
	return &inputAliasOutputConfirmed{
		aliasOutput: aliasOutput,
	}
}

func (inp *inputAliasOutputConfirmed) String() string {
	return fmt.Sprintf("{cmtLog.inputAliasOutputConfirmed, %v}", inp.aliasOutput)
}
