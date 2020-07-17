package kv

import (
	"testing"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/stretchr/testify/assert"
)

func TestBufferedKVStore(t *testing.T) {
	db := mapdb.NewMapDB()
	db.Set([]byte("abcd"), []byte("v1"))

	realm := db.WithRealm([]byte("ab"))

	v, err := realm.Get([]byte("cd"))
	assert.NoError(t, err)
	assert.Equal(t, []byte("v1"), v)

	b := NewBufferedKVStore(func() kvstore.KVStore { return realm })

	v, err = b.Get(Key([]byte("cd")))
	assert.NoError(t, err)
	assert.Equal(t, []byte("v1"), v)

	m := b.DangerouslyDumpToMap()
	assert.Equal(
		t,
		map[Key][]byte{
			Key([]byte("cd")): []byte("v1"),
		},
		m,
	)

	b.Set(Key([]byte("cd")), []byte("v2"))

	// not committed to DB
	v, err = realm.Get([]byte("cd"))
	assert.NoError(t, err)
	assert.Equal(t, []byte("v1"), v)

	v, err = b.Get(Key([]byte("cd")))
	assert.NoError(t, err)
	assert.Equal(t, []byte("v2"), v)

	m = b.DangerouslyDumpToMap()
	assert.Equal(
		t,
		map[Key][]byte{
			Key([]byte("cd")): []byte("v2"),
		},
		m,
	)
}
