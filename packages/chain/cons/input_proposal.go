// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/types"
)

// That's the main/initial input for the consensus.
type inputProposal struct {
	baseAliasOutput *types.Anchor
}

func NewInputProposal(baseAliasOutput *types.Anchor) gpa.Input {
	return &inputProposal{baseAliasOutput: baseAliasOutput}
}

func (ip *inputProposal) String() string {
	/*l1Commitment, err := transaction.L1CommitmentFromAliasOutput(ip.baseAliasOutput.GetAliasOutput())
	if err != nil {
		panic(fmt.Errorf("cannot extract L1 commitment from alias output: %w", err))
	}
	return fmt.Sprintf("{cons.inputProposal: baseAliasOutput=%v, l1Commitment=%v}", ip.baseAliasOutput, l1Commitment)*/
	return fmt.Sprintf("{cons.inputProposal: baseAliasOutput=%v}", ip.baseAliasOutput)
}
