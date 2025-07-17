package statetest

import (
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/samber/lo"
)

var TestL1Commitment = lo.Must(state.NewL1CommitmentFromBytes(testutil.TestBytes(state.L1CommitmentSize)))
