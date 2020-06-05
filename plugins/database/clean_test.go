package database

import (
	"github.com/iotaledger/goshimmer/packages/database"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDbClean(t *testing.T) {
	tmpdb, err := database.NewDB("tmp")
	assert.NoError(t, err)

	storeTmp := tmpdb.NewStore().WithRealm([]byte("realm"))

	err = storeTmp.Clear()
	assert.NoError(t, err)

	num := count(t, storeTmp)

	assert.NoError(t, err)
	assert.EqualValues(t, 0, num)

	err = storeTmp.Set([]byte("1"), []byte("a"))
	assert.NoError(t, err)
	err = storeTmp.Set([]byte("2"), []byte("b"))
	assert.NoError(t, err)
	err = storeTmp.Set([]byte("3"), []byte("c"))
	assert.NoError(t, err)

	num = count(t, storeTmp)

	assert.NoError(t, err)
	assert.EqualValues(t, 3, num)

	err = storeTmp.Clear()
	assert.NoError(t, err)

	num = count(t, storeTmp)

	assert.NoError(t, err)
	assert.EqualValues(t, 0, num)
}

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
