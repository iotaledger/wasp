package database

import (
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMapDB(t *testing.T) {
	db := mapdb.NewMapDB()

	realm := kvstore.Realm("really?")
	r1 := db.WithRealm(realm)

	key := []byte("key")
	value := []byte("value")
	err := r1.Set(key, value)
	assert.NoError(t, err)

	r2 := db.WithRealm(realm)

	musthave, err := r2.Has(key)
	assert.NoError(t, err)
	assert.True(t, musthave)
}
