package database

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/kvstore"
	"github.com/iotaledger/wasp/packages/kvstore/mapdb"
)

func TestMapDB(t *testing.T) {
	db := mapdb.NewMapDB()

	realm := kvstore.Realm("really?")
	r1, err := db.WithRealm(realm)
	require.NoError(t, err)

	key := []byte("key")
	value := []byte("value")
	err = r1.Set(key, value)
	require.NoError(t, err)

	r2, err := db.WithRealm(realm)
	require.NoError(t, err)

	musthave, err := r2.Has(key)
	require.NoError(t, err)
	require.True(t, musthave)
}
