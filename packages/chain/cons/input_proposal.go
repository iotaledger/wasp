// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
)

// That's the main/initial input for the consensus.
type inputProposal struct {
	baseAliasOutput *isc.StateAnchor
}

func NewInputProposal(baseAliasOutput *isc.StateAnchor) gpa.Input {
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
