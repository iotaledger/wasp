package origin_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/origin"
	"github.com/iotaledger/wasp/packages/state"
)

func TestOrigin(t *testing.T) {
	store := origin.InitChain(state.NewStore(mapdb.NewMapDB()), nil, 0)
	l1commitment := origin.L1Commitment(nil, 0)
	block, err := store.LatestBlock()
	require.NoError(t, err)
	require.True(t, l1commitment.Equals(block.L1Commitment()))
}
