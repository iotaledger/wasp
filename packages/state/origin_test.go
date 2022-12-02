package state

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/core/kvstore/mapdb"
)

func TestOrigin(t *testing.T) {
	store := InitChainStore(mapdb.NewMapDB())
	l1commitment := OriginL1Commitment()
	require.True(t, l1commitment.Equals(store.LatestBlock().L1Commitment()))
}
