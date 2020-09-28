package kv

import (
	"testing"

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

	b := NewBufferedKVStore(realm)

	v, err = b.Get(Key([]byte("cd")))
	assert.NoError(t, err)
	assert.Equal(t, []byte("v1"), v)

	m := b.DangerouslyDumpToMap().ToGoMap()
	assert.Equal(
		t,
		map[Key][]byte{
			Key([]byte("cd")): []byte("v1"),
		},
		m,
	)

	n := 0
	b.Iterate(EmptyPrefix, func(key Key, value []byte) bool {
		assert.Equal(t, Key([]byte("cd")), key)
		assert.Equal(t, []byte("v1"), value)
		n++
		return true
	})
	assert.Equal(t, 1, n)

	n = 0
	b.IterateKeys(EmptyPrefix, func(key Key) bool {
		assert.Equal(t, Key([]byte("cd")), key)
		n++
		return true
	})
	assert.Equal(t, 1, n)

	b.Set(Key([]byte("cd")), []byte("v2"))

	// not committed to DB
	v, err = realm.Get([]byte("cd"))
	assert.NoError(t, err)
	assert.Equal(t, []byte("v1"), v)

	v, err = b.Get(Key([]byte("cd")))
	assert.NoError(t, err)
	assert.Equal(t, []byte("v2"), v)

	m = b.DangerouslyDumpToMap().ToGoMap()
	assert.Equal(
		t,
		map[Key][]byte{
			Key([]byte("cd")): []byte("v2"),
		},
		m,
	)
}
