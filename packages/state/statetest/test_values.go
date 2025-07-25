package statetest

import (
	"github.com/iotaledger/wasp/v2/packages/state"
	"github.com/iotaledger/wasp/v2/packages/testutil/testval"
	"github.com/samber/lo"
)

var TestL1Commitment = lo.Must(state.NewL1CommitmentFromBytes(testval.TestBytes(state.L1CommitmentSize)))
