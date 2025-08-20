package database

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/database"
	"github.com/iotaledger/wasp/v2/packages/kvstore"
)

func count(t *testing.T, store kvstore.KVStore) int {
	ret := 0
	err := store.Iterate(kvstore.EmptyPrefix, func(k kvstore.Key, v kvstore.Value) bool {
		ret++
		t.Logf("key = %s value = %s", string(k), string(v))
		return true
	})
	if err != nil {
		panic(err)
	}
	return ret
}

func TestDbClean(t *testing.T) {
	tmpdb, err := database.NewDatabaseInMemory()
	require.NoError(t, err)

	storeTmp := tmpdb.KVStore()

	err = storeTmp.Clear()
	require.NoError(t, err)

	num := count(t, storeTmp)

	require.NoError(t, err)
	require.EqualValues(t, 0, num)

	err = storeTmp.Set([]byte("1"), []byte("a"))
	require.NoError(t, err)
	err = storeTmp.Set([]byte("2"), []byte("b"))
	require.NoError(t, err)
	err = storeTmp.Set([]byte("3"), []byte("c"))
	require.NoError(t, err)

	num = count(t, storeTmp)

	require.NoError(t, err)
	require.EqualValues(t, 3, num)

	err = storeTmp.Clear()
	require.NoError(t, err)

	num = count(t, storeTmp)

	require.NoError(t, err)
	require.EqualValues(t, 0, num)
}
