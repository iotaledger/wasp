package state

import (
	"testing"
	"time"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
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
		chainID := iscp.ChainIDFromAliasID(ledgerstate.NewAliasAddress([]byte("dummy")))
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
		require.EqualValues(t, 0, z.BlockIndex())
		require.True(t, z.Timestamp().IsZero())
		require.EqualValues(t, hashing.NilHash, z.PreviousStateHash())
		require.EqualValues(t, calcOriginStateHash(), z.StateCommitment())
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

		require.EqualValues(t, vs1.Copy().StateCommitment(), vs2.Copy().StateCommitment())
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

		currentTime := time.Now()
		su := NewStateUpdateWithBlocklogValues(1, currentTime, hashing.NilHash)
		su.Mutations().Set("key", []byte("value"))
		block1, err := newBlock(su.Mutations())
		require.NoError(t, err)

		err = vs1.ApplyBlock(block1)
		require.NoError(t, err)
		require.EqualValues(t, 1, vs1.BlockIndex())
		require.True(t, currentTime.Equal(vs1.Timestamp()))

		err = vs1.Commit(block1)
		require.NoError(t, err)
		require.EqualValues(t, 1, vs1.BlockIndex())
		require.True(t, currentTime.Equal(vs1.Timestamp()))

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
	t.Run("apply block after loading", func(t *testing.T) {
		store := mapdb.NewMapDB()
		chainID := iscp.RandomChainID([]byte("1"))
		_, exists, err := LoadSolidState(store, chainID)
		require.NoError(t, err)
		require.False(t, exists)

		vsOrig, err := CreateOriginState(store, chainID)
		require.NoError(t, err)

		time1 := time.Now()
		su := NewStateUpdateWithBlocklogValues(1, time1, hashing.NilHash)
		su.Mutations().Set("key", []byte("value"))
		block1, err := newBlock(su.Mutations())
		require.NoError(t, err)

		err = vsOrig.ApplyBlock(block1)
		require.NoError(t, err)
		require.EqualValues(t, 1, vsOrig.BlockIndex())
		require.True(t, time1.Equal(vsOrig.Timestamp()))

		time2 := time.Now()
		su = NewStateUpdateWithBlocklogValues(2, time2, vsOrig.PreviousStateHash())
		su.Mutations().Set("other_key", []byte("other_value"))
		block2, err := newBlock(su.Mutations())
		require.NoError(t, err)

		err = vsOrig.ApplyBlock(block2)
		require.NoError(t, err)
		require.EqualValues(t, 2, vsOrig.BlockIndex())
		require.True(t, time2.Equal(vsOrig.Timestamp()))

		err = vsOrig.Commit(block1, block2)
		require.NoError(t, err)
		require.EqualValues(t, 2, vsOrig.BlockIndex())
		require.True(t, time2.Equal(vsOrig.Timestamp()))

		vsLoaded, exists, err := LoadSolidState(store, chainID)
		require.NoError(t, err)
		require.True(t, exists)

		require.EqualValues(t, vsOrig.StateCommitment(), vsLoaded.StateCommitment())
		require.EqualValues(t, vsOrig.BlockIndex(), vsLoaded.BlockIndex())
		require.EqualValues(t, vsOrig.Timestamp(), vsLoaded.Timestamp())
		require.EqualValues(t, 2, vsLoaded.BlockIndex())

		time3 := time.Now()
		su = NewStateUpdateWithBlocklogValues(3, time3, vsLoaded.PreviousStateHash())
		su.Mutations().Set("more_keys", []byte("more_values"))
		block3, err := newBlock(su.Mutations())
		require.NoError(t, err)

		err = vsOrig.ApplyBlock(block3)
		require.NoError(t, err)
		require.EqualValues(t, 3, vsOrig.BlockIndex())
		require.True(t, time3.Equal(vsOrig.Timestamp()))

		err = vsLoaded.ApplyBlock(block3)
		require.NoError(t, err)
		require.EqualValues(t, 3, vsLoaded.BlockIndex())
		require.True(t, time3.Equal(vsLoaded.Timestamp()))

		require.EqualValues(t, vsOrig.StateCommitment(), vsLoaded.StateCommitment())
	})
	t.Run("state reader", func(t *testing.T) {
		store := mapdb.NewMapDB()
		chainID := iscp.RandomChainID([]byte("1"))
		_, exists, err := LoadSolidState(store, chainID)
		require.NoError(t, err)
		require.False(t, exists)

		vs1, err := CreateOriginState(store, chainID)
		require.NoError(t, err)

		currentTime := time.Now()
		su := NewStateUpdateWithBlocklogValues(1, currentTime, hashing.NilHash)
		su.Mutations().Set("key", []byte("value"))
		block1, err := newBlock(su.Mutations())
		require.NoError(t, err)

		err = vs1.ApplyBlock(block1)
		require.NoError(t, err)
		require.EqualValues(t, 1, vs1.BlockIndex())
		require.True(t, currentTime.Equal(vs1.Timestamp()))

		err = vs1.Commit()
		require.NoError(t, err)
		require.EqualValues(t, 1, vs1.BlockIndex())
		require.True(t, currentTime.Equal(vs1.Timestamp()))

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
		require.EqualValues(t, err, coreutil.ErrorStateInvalidated)
	})
}

func TestVariableStateBasic(t *testing.T) {
	chainID := iscp.ChainIDFromAliasID(ledgerstate.NewAliasAddress([]byte("dummy")))
	vs1, err := CreateOriginState(mapdb.NewMapDB(), chainID)
	require.NoError(t, err)
	h1 := vs1.StateCommitment()
	require.EqualValues(t, OriginStateHash(), h1)

	vs2 := vs1.Copy()
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

	vs3 := vs1.Copy()
	vs4 := vs2.Copy()

	require.EqualValues(t, vs3.StateCommitment(), vs4.StateCommitment())
}

func TestStateCommitmentAssociativity(t *testing.T) {
	store1 := mapdb.NewMapDB()
	store2 := mapdb.NewMapDB()
	chainID := iscp.RandomChainID([]byte("associative"))

	// vsNode1 index 0 vsNode2 index 0

	vsNode1, err := CreateOriginState(store1, chainID)
	require.NoError(t, err)
	vsNode2, err := CreateOriginState(store2, chainID)
	require.NoError(t, err)

	// vsNode1 index 1 vsNode2 index 0

	currentTime := time.Now()
	su := NewStateUpdateWithBlocklogValues(1, currentTime, hashing.NilHash)
	su.Mutations().Set("key", []byte("value"))
	block1, err := newBlock(su.Mutations())
	require.NoError(t, err)

	err = vsNode1.ApplyBlock(block1)
	require.NoError(t, err)
	require.EqualValues(t, 1, vsNode1.BlockIndex())
	require.True(t, currentTime.Equal(vsNode1.Timestamp()))
	sc1Node1BeforeCommit := vsNode1.StateCommitment()

	err = vsNode1.Commit(block1)
	require.NoError(t, err)
	require.EqualValues(t, 1, vsNode1.BlockIndex())
	require.True(t, currentTime.Equal(vsNode1.Timestamp()))
	sc1Node1AfterCommit := vsNode1.StateCommitment()
	require.Equal(t, sc1Node1BeforeCommit, sc1Node1AfterCommit)

	// vsNode1 index 2 vsNode2 index 0

	currentTime = time.Now()
	su = NewStateUpdateWithBlocklogValues(2, currentTime, vsNode1.PreviousStateHash())
	su.Mutations().Set("otherKey", []byte("otherValue"))
	block2, err := newBlock(su.Mutations())
	require.NoError(t, err)

	err = vsNode1.ApplyBlock(block2)
	require.NoError(t, err)
	require.EqualValues(t, 2, vsNode1.BlockIndex())
	require.True(t, currentTime.Equal(vsNode1.Timestamp()))
	sc2Node1BeforeCommit := vsNode1.StateCommitment()

	// vsNode1 index 2 vsNode2 index 2

	err = vsNode2.ApplyBlock(block1)
	require.NoError(t, err)
	err = vsNode2.ApplyBlock(block2)
	require.NoError(t, err)
	require.EqualValues(t, 2, vsNode2.BlockIndex())
	require.True(t, currentTime.Equal(vsNode2.Timestamp()))
	sc2Node2BeforeCommit := vsNode2.StateCommitment()
	require.Equal(t, sc2Node1BeforeCommit, sc2Node2BeforeCommit)

	err = vsNode1.Commit(block2)
	require.NoError(t, err)
	require.EqualValues(t, 2, vsNode1.BlockIndex())
	require.True(t, currentTime.Equal(vsNode1.Timestamp()))
	sc2Node1AfterCommit := vsNode1.StateCommitment()
	require.Equal(t, sc2Node1BeforeCommit, sc2Node1AfterCommit)
	require.Equal(t, sc2Node1AfterCommit, sc2Node2BeforeCommit)

	err = vsNode2.Commit(block1)
	require.NoError(t, err)
	err = vsNode2.Commit(block2)
	require.NoError(t, err)
	require.EqualValues(t, 2, vsNode2.BlockIndex())
	require.True(t, currentTime.Equal(vsNode2.Timestamp()))
	sc2Node2AfterCommit := vsNode2.StateCommitment()
	require.Equal(t, sc2Node2BeforeCommit, sc2Node2AfterCommit)
	require.Equal(t, sc2Node1AfterCommit, sc2Node2AfterCommit)
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

func TestVirtualStateMustOptimistic1(t *testing.T) {
	db := mapdb.NewMapDB()
	glb := coreutil.NewChainStateSync()
	glb.SetSolidIndex(0)
	baseline := glb.GetSolidIndexBaseline()
	chainID := iscp.RandomChainID([]byte("1"))
	vs, err := CreateOriginState(db, chainID)
	require.NoError(t, err)

	vsOpt := WrapMustOptimisticVirtualStateAccess(vs, baseline)

	h1 := vsOpt.StateCommitment()
	require.EqualValues(t, OriginStateHash(), h1)
	require.EqualValues(t, 0, vsOpt.BlockIndex())

	glb.InvalidateSolidIndex()
	require.PanicsWithValue(t, coreutil.ErrorStateInvalidated, func() {
		_ = vsOpt.StateCommitment()
	})
	require.PanicsWithValue(t, coreutil.ErrorStateInvalidated, func() {
		_ = vsOpt.BlockIndex()
	})
	require.PanicsWithValue(t, coreutil.ErrorStateInvalidated, func() {
		_, _ = vsOpt.ExtractBlock()
	})
	require.PanicsWithValue(t, coreutil.ErrorStateInvalidated, func() {
		_ = vsOpt.PreviousStateHash()
	})
	require.PanicsWithValue(t, coreutil.ErrorStateInvalidated, func() {
		_ = vsOpt.KVStore()
	})
}

func TestVirtualStateMustOptimistic2(t *testing.T) {
	db := mapdb.NewMapDB()
	glb := coreutil.NewChainStateSync()
	glb.SetSolidIndex(0)
	baseline := glb.GetSolidIndexBaseline()
	chainID := iscp.RandomChainID([]byte("1"))
	vs, err := CreateOriginState(db, chainID)
	require.NoError(t, err)

	vsOpt := WrapMustOptimisticVirtualStateAccess(vs, baseline)

	hash := vs.StateCommitment()
	hashOpt := vsOpt.StateCommitment()
	require.EqualValues(t, hash, hashOpt)

	hashPrev := hash
	upd := NewStateUpdateWithBlocklogValues(vsOpt.BlockIndex()+1, vsOpt.Timestamp().Add(1*time.Second), vsOpt.PreviousStateHash())
	vsOpt.ApplyStateUpdate(upd)
	hash = vs.StateCommitment()
	hashOpt = vsOpt.StateCommitment()
	require.EqualValues(t, hash, hashOpt)
	require.NotEqualValues(t, hashPrev, hashOpt)
}
