// Package statetest provides testing utilities for state package
package statetest

import (
	"math"
	"math/rand"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/kvstore/mapdb"
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

// RandomBlock is a test only function
func RandomBlock() state.Block {
	store := NewStoreWithUniqueWriteMutex(mapdb.NewMapDB())
	draft := store.NewOriginStateDraft()
	for i := 0; i < 3; i++ {
		draft.Set(kv.Key([]byte{byte(rand.Intn(math.MaxInt8))}), []byte{byte(rand.Intn(math.MaxInt8))})
	}

	return lo.Return1(store.Commit(draft))
}
