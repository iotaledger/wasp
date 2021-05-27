package state

import (
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/database/dbprovider"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/stretchr/testify/require"
)

func TestVirtualStateBasic(t *testing.T) {
	t.Run("create new1", func(t *testing.T) {
		db := mapdb.NewMapDB()
		vs1 := newVirtualState(db, nil)
		require.EqualValues(t, hashing.NilHash, vs1.Hash())
		require.Panics(t, func() {
			vs1.BlockIndex()
		})
	})
	t.Run("create new2", func(t *testing.T) {
		db := mapdb.NewMapDB()
		chainID := coretypes.NewChainID(ledgerstate.NewAliasAddress([]byte("dummy")))
		vs1 := newVirtualState(db, chainID)
		h1 := vs1.Hash()
		require.EqualValues(t, hashing.NilHash, h1)
		require.Panics(t, func() {
			vs1.BlockIndex()
		})
	})
	t.Run("zero state", func(t *testing.T) {
		db := mapdb.NewMapDB()
		vs1, blk := newZeroVirtualState(db, nil)
		h1 := vs1.Hash()
		require.EqualValues(t, OriginStateHash(), h1)
		require.EqualValues(t, 0, vs1.BlockIndex())
		require.EqualValues(t, NewOriginBlock().Bytes(), blk.Bytes())
	})
}

func TestOriginHashes(t *testing.T) {
	t.Run("origin state hash consistency ", func(t *testing.T) {
		t.Logf("origin state hash calculated: %s", calcOriginStateHash().String())
		require.EqualValues(t, OriginStateHashBase58, OriginStateHash().String())
		require.EqualValues(t, OriginStateHash().String(), calcOriginStateHash().String())
	})
	t.Run("zero state hash == origin state hash", func(t *testing.T) {
		z, _ := newZeroVirtualState(mapdb.NewMapDB(), nil)
		t.Logf("zero state hash = %s", z.Hash().String())
		require.EqualValues(t, calcOriginStateHash(), z.Hash())
	})
	t.Run("origin state construct", func(t *testing.T) {
		origBlock := NewOriginBlock()
		emptyState := newVirtualState(mapdb.NewMapDB(), nil)
		err := emptyState.ApplyBlock(origBlock)
		require.NoError(t, err)
		require.EqualValues(t, emptyState.Hash(), calcOriginStateHash())
	})
}

func TestStateWithDB(t *testing.T) {
	t.Run("state not found", func(t *testing.T) {
		log := testlogger.NewLogger(t)
		store := dbprovider.NewInMemoryDBProvider(log).GetPartition((nil))
		chainID := coretypes.RandomChainID([]byte("1"))
		_, exists, err := LoadSolidState(store, chainID)
		require.NoError(t, err)
		require.False(t, exists)
	})
	t.Run("save zero state", func(t *testing.T) {
		log := testlogger.NewLogger(t)
		store := dbprovider.NewInMemoryDBProvider(log).GetPartition((nil))
		chainID := coretypes.RandomChainID([]byte("1"))
		_, exists, err := LoadSolidState(store, chainID)
		require.NoError(t, err)
		require.False(t, exists)

		vs1, err := CreateOriginState(store, chainID)
		require.NoError(t, err)
		require.EqualValues(t, 0, vs1.BlockIndex())
		require.True(t, vs1.Timestamp().IsZero())

		vs2, exists, err := LoadSolidState(store, chainID)
		require.NoError(t, err)
		require.True(t, exists)

		require.EqualValues(t, vs1.Hash(), vs2.Hash())
		require.EqualValues(t, vs1.BlockIndex(), vs2.BlockIndex())
		require.EqualValues(t, vs1.Timestamp(), vs2.Timestamp())
		require.True(t, vs2.Timestamp().IsZero())
		require.EqualValues(t, 0, vs2.BlockIndex())

		require.EqualValues(t, vs1.Clone().Hash(), vs2.Clone().Hash())
	})
	t.Run("load 0 block", func(t *testing.T) {
		log := testlogger.NewLogger(t)
		store := dbprovider.NewInMemoryDBProvider(log).GetPartition((nil))
		chainID := coretypes.RandomChainID([]byte("1"))
		_, exists, err := LoadSolidState(store, chainID)
		require.NoError(t, err)
		require.False(t, exists)

		vs1, err := CreateOriginState(store, chainID)
		require.NoError(t, err)
		require.EqualValues(t, 0, vs1.BlockIndex())
		require.True(t, vs1.Timestamp().IsZero())

		data, err := LoadBlockBytes(store, 0)
		require.NoError(t, err)
		require.EqualValues(t, NewOriginBlock().Bytes(), data)
	})
	t.Run("apply, save and load block 1", func(t *testing.T) {
		log := testlogger.NewLogger(t)
		store := dbprovider.NewInMemoryDBProvider(log).GetPartition((nil))
		chainID := coretypes.RandomChainID([]byte("1"))
		_, exists, err := LoadSolidState(store, chainID)
		require.NoError(t, err)
		require.False(t, exists)

		vs1, err := CreateOriginState(store, chainID)
		require.NoError(t, err)

		nowis := time.Now()
		su := NewStateUpdateWithBlockIndexMutation(1)
		su1 := NewStateUpdate(nowis)
		su1.Mutations().Set("key", []byte("value"))
		block1, err := NewBlock(su, su1)
		require.NoError(t, err)

		err = vs1.ApplyBlock(block1)
		require.NoError(t, err)
		require.EqualValues(t, 1, vs1.BlockIndex())
		require.True(t, nowis.Equal(vs1.Timestamp()))

		err = vs1.Commit(block1)
		require.NoError(t, err)
		require.EqualValues(t, 1, vs1.BlockIndex())
		require.True(t, nowis.Equal(vs1.Timestamp()))

		vs2, exists, err := LoadSolidState(store, chainID)
		require.NoError(t, err)
		require.True(t, exists)

		require.EqualValues(t, vs1.Hash(), vs2.Hash())
		require.EqualValues(t, vs1.BlockIndex(), vs2.BlockIndex())
		require.EqualValues(t, vs1.Timestamp(), vs2.Timestamp())
		require.EqualValues(t, 1, vs2.BlockIndex())

		data, err := LoadBlockBytes(store, 0)
		require.NoError(t, err)
		require.EqualValues(t, NewOriginBlock().Bytes(), data)

		data, err = LoadBlockBytes(store, 1)
		require.NoError(t, err)
		require.EqualValues(t, block1.Bytes(), data)

		data = vs2.KVStoreReader().MustGet("key")
		require.EqualValues(t, []byte("value"), data)

		require.EqualValues(t, vs1.Hash(), vs2.Hash())
	})
	t.Run("state reader", func(t *testing.T) {
		log := testlogger.NewLogger(t)
		store := dbprovider.NewInMemoryDBProvider(log).GetPartition((nil))
		chainID := coretypes.RandomChainID([]byte("1"))
		_, exists, err := LoadSolidState(store, chainID)
		require.NoError(t, err)
		require.False(t, exists)

		vs1, err := CreateOriginState(store, chainID)
		require.NoError(t, err)

		nowis := time.Now()
		su := NewStateUpdateWithBlockIndexMutation(1)
		su1 := NewStateUpdate(nowis)
		su1.Mutations().Set("key", []byte("value"))
		block1, err := NewBlock(su, su1)
		require.NoError(t, err)

		err = vs1.ApplyBlock(block1)
		require.NoError(t, err)
		require.EqualValues(t, 1, vs1.BlockIndex())
		require.True(t, nowis.Equal(vs1.Timestamp()))

		err = vs1.Commit()
		require.NoError(t, err)
		require.EqualValues(t, 1, vs1.BlockIndex())
		require.True(t, nowis.Equal(vs1.Timestamp()))

		vs2, exists, err := LoadSolidState(store, chainID)
		require.NoError(t, err)
		require.True(t, exists)

		rdr, err := NewStateReader(store, chainID)
		require.NoError(t, err)

		require.EqualValues(t, vs2.BlockIndex(), rdr.BlockIndex())
		require.EqualValues(t, vs2.Timestamp(), rdr.Timestamp())
		require.EqualValues(t, vs2.Hash().String(), rdr.Hash().String())
		require.EqualValues(t, "value", string(rdr.KVStoreReader().MustGet("key")))
	})
}

func TestVariableStateBasic(t *testing.T) {
	chainID := coretypes.NewChainID(ledgerstate.NewAliasAddress([]byte("dummy")))
	store := dbprovider.NewInMemoryDBProvider(testlogger.NewLogger(t)).GetPartition((nil))
	vs1, err := CreateOriginState(store, chainID)
	require.NoError(t, err)
	h1 := vs1.Hash()
	require.EqualValues(t, OriginStateHash(), h1)

	vs2 := vs1.Clone()
	h2 := vs2.Hash()
	require.EqualValues(t, h1, h2)

	vs1.KVStore().Set(kv.Key(coreutil.StatePrefixBlockIndex), codec.EncodeUint64(1))
	vs1.KVStore().Set("num", codec.EncodeInt64(int64(123)))
	vs1.KVStore().Set("kuku", codec.EncodeString("A"))
	vs1.KVStore().Set("mumu", codec.EncodeString("B"))

	vs2.KVStore().Set(kv.Key(coreutil.StatePrefixBlockIndex), codec.EncodeUint64(1))
	vs2.KVStore().Set("mumu", codec.EncodeString("B"))
	vs2.KVStore().Set("kuku", codec.EncodeString("A"))
	vs2.KVStore().Set("num", codec.EncodeInt64(int64(123)))

	require.EqualValues(t, vs1.Hash(), vs2.Hash())

	vs3 := vs1.Clone()
	vs4 := vs2.Clone()

	require.EqualValues(t, vs3.Hash(), vs4.Hash())
}

func TestStateReader(t *testing.T) {
	t.Run("state not found", func(t *testing.T) {
		log := testlogger.NewLogger(t)
		store := dbprovider.NewInMemoryDBProvider(log).GetPartition((nil))
		chainID := coretypes.RandomChainID([]byte("1"))
		st, err := NewStateReader(store, chainID)
		require.NoError(t, err)
		ok, err := st.KVStoreReader().Has("kuku")
		require.NoError(t, err)
		require.False(t, ok)
	})

}
