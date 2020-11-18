package buffered

import (
	"testing"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/kv"
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

	v, err = b.Get(kv.Key([]byte("cd")))
	assert.NoError(t, err)
	assert.Equal(t, []byte("v1"), v)

	m := kv.ToGoMap(b.DangerouslyDumpToDict())
	assert.Equal(
		t,
		map[kv.Key][]byte{
			kv.Key([]byte("cd")): []byte("v1"),
		},
		m,
	)

	n := 0
	b.Iterate(kv.EmptyPrefix, func(key kv.Key, value []byte) bool {
		assert.Equal(t, kv.Key([]byte("cd")), key)
		assert.Equal(t, []byte("v1"), value)
		n++
		return true
	})
	assert.Equal(t, 1, n)

	n = 0
	b.IterateKeys(kv.EmptyPrefix, func(key kv.Key) bool {
		assert.Equal(t, kv.Key([]byte("cd")), key)
		n++
		return true
	})
	assert.Equal(t, 1, n)

	b.Set(kv.Key([]byte("cd")), []byte("v2"))

	// not committed to DB
	v, err = realm.Get([]byte("cd"))
	assert.NoError(t, err)
	assert.Equal(t, []byte("v1"), v)

	v, err = b.Get(kv.Key([]byte("cd")))
	assert.NoError(t, err)
	assert.Equal(t, []byte("v2"), v)

	m = kv.ToGoMap(b.DangerouslyDumpToDict())
	assert.Equal(
		t,
		map[kv.Key][]byte{
			kv.Key([]byte("cd")): []byte("v2"),
		},
		m,
	)
}
