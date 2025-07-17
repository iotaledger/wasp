package statetest

import (
	"github.com/iotaledger/wasp/v2/packages/state"
	"github.com/iotaledger/wasp/v2/packages/testutil/testval"
	"github.com/samber/lo"
)

var TestL1Commitment = lo.Must(state.NewL1CommitmentFromBytes(testval.TestBytes(state.L1CommitmentSize)))

var TestBlock = func() state.Block {
	store := state.NewStoreWithUniqueWriteMutex(mapdb.NewMapDB())
	draft := store.NewOriginStateDraft()
	for i := 0; i < 3; i++ {
		draft.Set(kv.Key([]byte{byte((i + 1) * 6973)}), []byte{byte((i + 1) * 9137)})
	}

	block, _ := store.Commit(draft)
	return block
}()
