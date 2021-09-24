package state

import (
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/optimism"
	"github.com/stretchr/testify/require"
)

func TestVirtualStateBasic(t *testing.T) {
	t.Run("create new1", func(t *testing.T) {
		db := mapdb.NewMapDB()
		vs1 := newVirtualState(db, nil)
		require.Panics(t, func() {
			vs1.BlockIndex()
		})
	})
	t.Run("create new2", func(t *testing.T) {
		db := mapdb.NewMapDB()
		chainID := iscp.NewChainID(ledgerstate.NewAliasAddress([]byte("dummy")))
		vs1 := newVirtualState(db, chainID)
		require.Panics(t, func() {
			vs1.BlockIndex()
		})
	})
	t.Run("zero state", func(t *testing.T) {
		db := mapdb.NewMapDB()
		vs1, blk := newZeroVirtualState(db, nil)
		h1 := vs1.StateCommitment()
		require.EqualValues(t, OriginStateHash(), h1)
		require.EqualValues(t, 0, vs1.BlockIndex())
		require.EqualValues(t, newOriginBlock().Bytes(), blk.Bytes())
	})
}

func TestOriginHashes(t *testing.T) {
	t.Run("origin state hash consistency ", func(t *testing.T) {
		t.Logf("origin state hash calculated: %s", calcOriginStateHash().String())
		require.EqualValues(t, OriginStateHashBase58, OriginStateHash().String())
		require.EqualValues(t, OriginStateHash().String(), calcOriginStateHash().String())
	})
	t.Run("zero state hash == origin state hash", func(t *testing.T) {
		z, origBlock := newZeroVirtualState(mapdb.NewMapDB(), nil)
		require.EqualValues(t, 0, origBlock.BlockIndex())
		require.True(t, origBlock.Timestamp().IsZero())
		require.EqualValues(t, hashing.NilHash, origBlock.PreviousStateHash())
		t.Logf("zero state hash = %s", z.StateCommitment().String())
		require.EqualValues(t, calcOriginStateHash(), z.StateCommitment())
	})
	t.Run("origin state construct", func(t *testing.T) {
		origBlock := newOriginBlock()
		require.EqualValues(t, 0, origBlock.BlockIndex())
		require.True(t, origBlock.Timestamp().IsZero())
		require.EqualValues(t, hashing.NilHash, origBlock.PreviousStateHash())

		emptyState := newVirtualState(mapdb.NewMapDB(), nil)
		err := emptyState.ApplyBlock(origBlock)
		require.NoError(t, err)
		require.EqualValues(t, emptyState.StateCommitment(), calcOriginStateHash())
		require.EqualValues(t, hashing.NilHash, emptyState.PreviousStateHash())
	})
}

func TestStateWithDB(t *testing.T) {
	t.Run("state not found", func(t *testing.T) {
		store := mapdb.NewMapDB()
		chainID := iscp.RandomChainID([]byte("1"))
		_, exists, err := LoadSolidState(store, chainID)
		require.NoError(t, err)
		require.False(t, exists)
	})
	t.Run("save zero state", func(t *testing.T) {
		store := mapdb.NewMapDB()
		chainID := iscp.RandomChainID([]byte("1"))
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

		require.EqualValues(t, vs1.StateCommitment(), vs2.StateCommitment())
		require.EqualValues(t, vs1.BlockIndex(), vs2.BlockIndex())
		require.EqualValues(t, vs1.Timestamp(), vs2.Timestamp())
		require.EqualValues(t, vs1.PreviousStateHash(), vs2.PreviousStateHash())
		require.True(t, vs2.Timestamp().IsZero())
		require.EqualValues(t, 0, vs2.BlockIndex())
		require.EqualValues(t, hashing.NilHash, vs2.PreviousStateHash())

		require.EqualValues(t, vs1.Clone().StateCommitment(), vs2.Clone().StateCommitment())
	})
	t.Run("load 0 block", func(t *testing.T) {
		store := mapdb.NewMapDB()
		chainID := iscp.RandomChainID([]byte("1"))
		_, exists, err := LoadSolidState(store, chainID)
		require.NoError(t, err)
		require.False(t, exists)

		vs1, err := CreateOriginState(store, chainID)
		require.NoError(t, err)
		require.EqualValues(t, 0, vs1.BlockIndex())
		require.True(t, vs1.Timestamp().IsZero())

		data, err := LoadBlockBytes(store, 0)
		require.NoError(t, err)
		require.EqualValues(t, newOriginBlock().Bytes(), data)
	})
	t.Run("apply, save and load block 1", func(t *testing.T) {
		store := mapdb.NewMapDB()
		chainID := iscp.RandomChainID([]byte("1"))
		_, exists, err := LoadSolidState(store, chainID)
		require.NoError(t, err)
		require.False(t, exists)

		vs1, err := CreateOriginState(store, chainID)
		require.NoError(t, err)

		nowis := time.Now()
		su := NewStateUpdateWithBlocklogValues(1, nowis, hashing.NilHash)
		su.Mutations().Set("key", []byte("value"))
		block1, err := newBlock(su)
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

		require.EqualValues(t, vs1.StateCommitment(), vs2.StateCommitment())
		require.EqualValues(t, vs1.BlockIndex(), vs2.BlockIndex())
		require.EqualValues(t, vs1.Timestamp(), vs2.Timestamp())
		require.EqualValues(t, 1, vs2.BlockIndex())

		data, err := LoadBlockBytes(store, 0)
		require.NoError(t, err)
		require.EqualValues(t, newOriginBlock().Bytes(), data)

		data, err = LoadBlockBytes(store, 1)
		require.NoError(t, err)
		require.EqualValues(t, block1.Bytes(), data)

		data = vs2.KVStoreReader().MustGet("key")
		require.EqualValues(t, []byte("value"), data)

		require.EqualValues(t, vs1.StateCommitment(), vs2.StateCommitment())
	})
	t.Run("state reader", func(t *testing.T) {
		store := mapdb.NewMapDB()
		chainID := iscp.RandomChainID([]byte("1"))
		_, exists, err := LoadSolidState(store, chainID)
		require.NoError(t, err)
		require.False(t, exists)

		vs1, err := CreateOriginState(store, chainID)
		require.NoError(t, err)

		nowis := time.Now()
		su := NewStateUpdateWithBlocklogValues(1, nowis, hashing.NilHash)
		su.Mutations().Set("key", []byte("value"))
		block1, err := newBlock(su)
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

		glb := coreutil.NewChainStateSync()
		glb.SetSolidIndex(0)
		rdr := NewOptimisticStateReader(store, glb)

		bi, err := rdr.BlockIndex()
		require.NoError(t, err)
		require.EqualValues(t, vs2.BlockIndex(), bi)

		ts, err := rdr.Timestamp()
		require.NoError(t, err)
		require.EqualValues(t, vs2.Timestamp(), ts)

		h, err := rdr.Hash()
		require.NoError(t, err)
		require.EqualValues(t, vs2.StateCommitment(), h)
		require.EqualValues(t, "value", string(rdr.KVStoreReader().MustGet("key")))

		glb.InvalidateSolidIndex()
		_, err = rdr.Hash()
		require.Error(t, err)
		require.EqualValues(t, err, optimism.ErrStateHasBeenInvalidated)
	})
}

func TestVariableStateBasic(t *testing.T) {
	chainID := iscp.NewChainID(ledgerstate.NewAliasAddress([]byte("dummy")))
	vs1, err := CreateOriginState(mapdb.NewMapDB(), chainID)
	require.NoError(t, err)
	h1 := vs1.StateCommitment()
	require.EqualValues(t, OriginStateHash(), h1)

	vs2 := vs1.Clone()
	h2 := vs2.StateCommitment()
	require.EqualValues(t, h1, h2)

	vs1.KVStore().Set(kv.Key(coreutil.StatePrefixBlockIndex), codec.EncodeUint64(1))
	vs1.KVStore().Set("num", codec.EncodeInt64(int64(123)))
	vs1.KVStore().Set("kuku", codec.EncodeString("A"))
	vs1.KVStore().Set("mumu", codec.EncodeString("B"))

	vs2.KVStore().Set(kv.Key(coreutil.StatePrefixBlockIndex), codec.EncodeUint64(1))
	vs2.KVStore().Set("mumu", codec.EncodeString("B"))
	vs2.KVStore().Set("kuku", codec.EncodeString("A"))
	vs2.KVStore().Set("num", codec.EncodeInt64(int64(123)))

	require.EqualValues(t, vs1.StateCommitment(), vs2.StateCommitment())

	vs3 := vs1.Clone()
	vs4 := vs2.Clone()

	require.EqualValues(t, vs3.StateCommitment(), vs4.StateCommitment())
}

func TestStateReader(t *testing.T) {
	t.Run("state not found", func(t *testing.T) {
		store := mapdb.NewMapDB()
		chainID := iscp.RandomChainID([]byte("1"))
		_, err := CreateOriginState(store, chainID)
		require.NoError(t, err)

		glb := coreutil.NewChainStateSync()
		glb.SetSolidIndex(0)
		st := NewOptimisticStateReader(store, glb)
		ok, err := st.KVStoreReader().Has("kuku")
		require.NoError(t, err)
		require.False(t, ok)
	})
}
