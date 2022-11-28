// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
)

// That's the main/initial input for the consensus.
type inputProposal struct {
	baseAliasOutput *isc.AliasOutputWithID
}

func NewInputProposal(baseAliasOutput *isc.AliasOutputWithID) gpa.Input {
	return &inputProposal{baseAliasOutput: baseAliasOutput}
}
