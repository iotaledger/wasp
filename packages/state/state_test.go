package state

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/trie"
	"github.com/iotaledger/wasp/packages/testutil/testmisc"
	"github.com/stretchr/testify/require"
)

func TestOriginHashes(t *testing.T) {
	t.Run("create new", func(t *testing.T) {
		vs1 := NewVirtualState(mapdb.NewMapDB())
		require.Panics(t, func() {
			vs1.BlockIndex()
		})
	})
	t.Run("origin state hash consistency ", func(t *testing.T) {
		t.Logf("origin state hash calculated: %s", calcOriginStateHash().String())
		require.EqualValues(t, OriginStateCommitmentHex, OriginStateCommitment().String())
		require.EqualValues(t, OriginStateCommitment().String(), calcOriginStateHash().String())
	})
	t.Run("zero state hash == origin state hash", func(t *testing.T) {
		z := NewVirtualState(mapdb.NewMapDB())
		require.Nil(t, trie.RootCommitment(z.TrieNodeStore()))
	})
	t.Run("create origin", func(t *testing.T) {
		chainID := testmisc.RandChainID()
		vs, err := CreateOriginState(mapdb.NewMapDB(), chainID)
		require.NoError(t, err)
		require.True(t, trie.EqualCommitments(trie.RootCommitment(vs.TrieNodeStore()), OriginStateCommitment()))
		require.EqualValues(t, calcOriginStateHash(), trie.RootCommitment(vs.TrieNodeStore()))
	})
}

func TestStateWithDB(t *testing.T) {
	t.Run("save state", func(t *testing.T) {
		chainID := testmisc.RandChainID()
		store := mapdb.NewMapDB()
		vs, err := CreateOriginState(store, chainID)
		require.NoError(t, err)
		vs.Commit()
		cc := trie.RootCommitment(vs.TrieNodeStore())
		err = vs.Save()
		require.NoError(t, err)
		cs := trie.RootCommitment(vs.TrieNodeStore())
		require.True(t, trie.EqualCommitments(cc, cs))
		_, exists, err := LoadSolidState(store, chainID)
		require.NoError(t, err)
		require.True(t, exists)
	})
	t.Run("state not found", func(t *testing.T) {
		store := mapdb.NewMapDB()
		chainID := iscp.RandomChainID([]byte("1"))
		_, exists, err := LoadSolidState(store, chainID)
		require.NoError(t, err)
		require.False(t, exists)
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
		su := NewStateUpdateWithBlockLogValues(1, currentTime, RandL1Commitment())
		su.Mutations().Set("key", []byte("value"))
		block1, err := newBlock(su.Mutations())
		require.NoError(t, err)

		err = vs1.ApplyBlock(block1)
		require.NoError(t, err)
		require.EqualValues(t, 1, vs1.BlockIndex())
		require.True(t, currentTime.Equal(vs1.Timestamp()))

		err = vs1.Save(block1)
		require.NoError(t, err)
		require.EqualValues(t, 1, vs1.BlockIndex())
		require.True(t, currentTime.Equal(vs1.Timestamp()))

		vs2, exists, err := LoadSolidState(store, chainID)
		require.NoError(t, err)
		require.True(t, exists)

		require.EqualValues(t, trie.RootCommitment(vs1.TrieNodeStore()), trie.RootCommitment(vs2.TrieNodeStore()))
		require.EqualValues(t, vs1.BlockIndex(), vs2.BlockIndex())
		require.EqualValues(t, vs1.Timestamp(), vs2.Timestamp())
		require.EqualValues(t, 1, vs2.BlockIndex())

		data, err := LoadBlockBytes(store, 0)
		require.NoError(t, err)
		// require.EqualValues(t, newBlock().Bytes(), data)

		data, err = LoadBlockBytes(store, 1)
		require.NoError(t, err)
		require.EqualValues(t, block1.Bytes(), data)

		data = vs2.KVStoreReader().MustGet("key")
		require.EqualValues(t, []byte("value"), data)

		require.EqualValues(t, trie.RootCommitment(vs1.TrieNodeStore()), trie.RootCommitment(vs2.TrieNodeStore()))
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
		su := NewStateUpdateWithBlockLogValues(1, time1, RandL1Commitment())
		su.Mutations().Set("key", []byte("value"))
		block1, err := newBlock(su.Mutations())
		require.NoError(t, err)

		err = vsOrig.ApplyBlock(block1)
		require.NoError(t, err)
		require.EqualValues(t, 1, vsOrig.BlockIndex())
		require.True(t, time1.Equal(vsOrig.Timestamp()))

		time2 := time.Now()
		su = NewStateUpdateWithBlockLogValues(2, time2, vsOrig.PreviousL1Commitment())
		su.Mutations().Set("other_key", []byte("other_value"))
		block2, err := newBlock(su.Mutations())
		require.NoError(t, err)

		err = vsOrig.ApplyBlock(block2)
		require.NoError(t, err)
		require.EqualValues(t, 2, vsOrig.BlockIndex())
		require.True(t, time2.Equal(vsOrig.Timestamp()))

		err = vsOrig.Save(block1, block2)
		require.NoError(t, err)
		require.EqualValues(t, 2, vsOrig.BlockIndex())
		require.True(t, time2.Equal(vsOrig.Timestamp()))

		vsLoaded, exists, err := LoadSolidState(store, chainID)
		require.NoError(t, err)
		require.True(t, exists)

		require.EqualValues(t, trie.RootCommitment(vsOrig.TrieNodeStore()), trie.RootCommitment(vsLoaded.TrieNodeStore()))
		require.EqualValues(t, vsOrig.BlockIndex(), vsLoaded.BlockIndex())
		require.EqualValues(t, vsOrig.Timestamp(), vsLoaded.Timestamp())
		require.EqualValues(t, 2, vsLoaded.BlockIndex())

		time3 := time.Now()
		su = NewStateUpdateWithBlockLogValues(3, time3, vsLoaded.PreviousL1Commitment())
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

		require.EqualValues(t, trie.RootCommitment(vsOrig.TrieNodeStore()), trie.RootCommitment(vsLoaded.TrieNodeStore()))
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
		su := NewStateUpdateWithBlockLogValues(1, currentTime, RandL1Commitment())
		su.Mutations().Set("key", []byte("value"))
		block1, err := newBlock(su.Mutations())
		require.NoError(t, err)

		err = vs1.ApplyBlock(block1)
		require.NoError(t, err)
		require.EqualValues(t, 1, vs1.BlockIndex())
		require.True(t, currentTime.Equal(vs1.Timestamp()))

		err = vs1.Save()
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

		_, err = rdr.KVStoreReader().Get("1")
		require.NoError(t, err)
		require.EqualValues(t, "value", string(rdr.KVStoreReader().MustGet("key")))

		glb.InvalidateSolidIndex()
		_, err = rdr.KVStoreReader().Get("1")
		require.Error(t, err)
		require.EqualValues(t, err, coreutil.ErrorStateInvalidated)
	})
}

func TestStateBasic(t *testing.T) {
	chainID := iscp.ChainIDFromAliasID(tpkg.RandAliasAddress().AliasID())
	vs1, err := CreateOriginState(mapdb.NewMapDB(), &chainID)
	require.NoError(t, err)
	h1 := trie.RootCommitment(vs1.TrieNodeStore())
	require.True(t, trie.EqualCommitments(OriginStateCommitment(), h1))

	vs2 := vs1.Copy()
	h2 := trie.RootCommitment(vs2.TrieNodeStore())
	require.EqualValues(t, h1, h2)

	vs1.KVStore().Set(kv.Key(coreutil.StatePrefixBlockIndex), codec.EncodeUint32(1))
	vs1.KVStore().Set("num", codec.EncodeInt64(int64(123)))
	vs1.KVStore().Set("kuku", codec.EncodeString("A"))
	vs1.KVStore().Set("mumu", codec.EncodeString("B"))

	vs2.KVStore().Set(kv.Key(coreutil.StatePrefixBlockIndex), codec.EncodeUint32(1))
	vs2.KVStore().Set("mumu", codec.EncodeString("B"))
	vs2.KVStore().Set("kuku", codec.EncodeString("A"))
	vs2.KVStore().Set("num", codec.EncodeInt64(int64(123)))

	require.EqualValues(t, trie.RootCommitment(vs1.TrieNodeStore()), trie.RootCommitment(vs2.TrieNodeStore()))

	vs3 := vs1.Copy()
	vs4 := vs2.Copy()

	require.EqualValues(t, trie.RootCommitment(vs3.TrieNodeStore()), trie.RootCommitment(vs4.TrieNodeStore()))
}

func TestStateReader(t *testing.T) {
	t.Run("state not found", func(t *testing.T) {
		store := mapdb.NewMapDB()
		chainID := iscp.RandomChainID([]byte("1"))
		os, err := CreateOriginState(store, chainID)
		require.NoError(t, err)
		err = os.Save()
		require.NoError(t, err)
		c1 := trie.RootCommitment(os.TrieNodeStore())

		glb := coreutil.NewChainStateSync()
		glb.SetSolidIndex(0)
		st := NewOptimisticStateReader(store, glb)
		ok, err := st.KVStoreReader().Has("kuku")
		require.NoError(t, err)
		require.False(t, ok)

		c2 := trie.RootCommitment(st.TrieNodeStore())
		require.True(t, trie.EqualCommitments(c1, c2))
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

	h1 := trie.RootCommitment(vsOpt.TrieNodeStore())
	require.True(t, trie.EqualCommitments(OriginStateCommitment(), h1))
	require.EqualValues(t, 0, vsOpt.BlockIndex())

	glb.InvalidateSolidIndex()
	require.PanicsWithValue(t, coreutil.ErrorStateInvalidated, func() {
		_ = trie.RootCommitment(vsOpt.TrieNodeStore())
	})
	require.PanicsWithValue(t, coreutil.ErrorStateInvalidated, func() {
		_ = vsOpt.BlockIndex()
	})
	require.PanicsWithValue(t, coreutil.ErrorStateInvalidated, func() {
		_, _ = vsOpt.ExtractBlock()
	})
	require.PanicsWithValue(t, coreutil.ErrorStateInvalidated, func() {
		_ = vsOpt.PreviousL1Commitment()
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

	hash := trie.RootCommitment(vs.TrieNodeStore())
	hashOpt := trie.RootCommitment(vsOpt.TrieNodeStore())
	require.EqualValues(t, hash, hashOpt)

	hashPrev := hash
	prev := NewL1Commitment(trie.RootCommitment(vsOpt.TrieNodeStore()), hashing.RandomHash(nil))
	upd := NewStateUpdateWithBlockLogValues(vsOpt.BlockIndex()+1, vsOpt.Timestamp().Add(1*time.Second), prev)
	vsOpt.ApplyStateUpdate(upd)
	hash = trie.RootCommitment(vs.TrieNodeStore())
	hashOpt = trie.RootCommitment(vsOpt.TrieNodeStore())
	require.EqualValues(t, hash, hashOpt)
	require.EqualValues(t, hashPrev, hashOpt)
}
