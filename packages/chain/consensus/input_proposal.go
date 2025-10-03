// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/gpa"
	"github.com/iotaledger/wasp/v2/packages/isc"
)

// That's the main/initial input for the consensus.
type inputProposal struct {
	baseAnchor *isc.StateAnchor
}

func NewInputProposal(baseAnchor *isc.StateAnchor) gpa.Input {
	return &inputProposal{baseAnchor: baseAnchor}
}

func (ip *inputProposal) String() string {
	/*l1Commitment, err := transaction.L1CommitmentFromAnchor(ip.baseAnchor.GetAnchor())
	if err != nil {
		panic(fmt.Errorf("cannot extract L1 commitment from alias output: %w", err))
	}
	return fmt.Sprintf("{cons.inputProposal: baseAnchor=%v, l1Commitment=%v}", ip.baseAnchor, l1Commitment)*/
	return fmt.Sprintf("{cons.inputProposal: baseAnchor=%v}", ip.baseAnchor)
}
