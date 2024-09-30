// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmt_log

import (
	"fmt"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/gpa"
)

type inputAliasOutputConfirmed struct {
	aliasOutput *iscmove.AnchorWithRef
}

func NewInputAliasOutputConfirmed(aliasOutput *iscmove.AnchorWithRef) gpa.Input {
	return &inputAliasOutputConfirmed{
		aliasOutput: aliasOutput,
	}
}

func (inp *inputAliasOutputConfirmed) String() string {
	return fmt.Sprintf("{cmtLog.inputAliasOutputConfirmed, %v}", inp.aliasOutput)
}
