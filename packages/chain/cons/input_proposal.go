// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cons

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/transaction"
)

// That's the main/initial input for the consensus.
type inputProposal struct {
	baseAnchorOutput *isc.AnchorOutputWithID
}

func NewInputProposal(baseAnchorOutput *isc.AnchorOutputWithID) gpa.Input {
	return &inputProposal{baseAnchorOutput: baseAnchorOutput}
}

func (ip *inputProposal) String() string {
	l1Commitment, err := transaction.L1CommitmentFromAnchorOutput(ip.baseAnchorOutput.GetAnchorOutput())
	if err != nil {
		panic(fmt.Errorf("cannot extract L1 commitment from alias output: %w", err))
	}
	return fmt.Sprintf("{cons.inputProposal: baseAnchorOutput=%v, l1Commitment=%v}", ip.baseAnchorOutput, l1Commitment)
}
