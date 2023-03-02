package state

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
)

func TestOrigin(t *testing.T) {
	store := InitChainStore(mapdb.NewMapDB())
	l1commitment := OriginL1Commitment()
	block, err := store.LatestBlock()
	require.NoError(t, err)
	require.True(t, l1commitment.Equals(block.L1Commitment()))
}
