// Package statetest provides testing utilities for state package
package statetest

import (
	"github.com/iotaledger/wasp/v2/packages/state"
	"github.com/iotaledger/wasp/v2/packages/util"
)

func NewRandL1Commitment() *state.L1Commitment {
	d := make([]byte, state.L1CommitmentSize)
	_, _ = util.NewPseudoRand().Read(d)
	ret, err := state.NewL1CommitmentFromBytes(d)
	if err != nil {
		panic(err)
	}
	return ret
}
