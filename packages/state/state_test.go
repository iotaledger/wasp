package state

import (
	"fmt"
	"math"
	"math/rand"
	"testing"
	"time"

	"github.com/iotaledger/hive.go/kvstore"
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
		vs1 := newVirtualState(mapdb.NewMapDB())
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
		z := newVirtualState(mapdb.NewMapDB())
		require.Nil(t, trie.RootCommitment(z.TrieAccess()))
	})
	t.Run("create origin", func(t *testing.T) {
		chainID := testmisc.RandChainID()
		vs, err := CreateOriginState(mapdb.NewMapDB(), chainID)
		require.NoError(t, err)
		require.True(t, trie.EqualCommitments(trie.RootCommitment(vs.TrieAccess()), OriginStateCommitment()))
		require.EqualValues(t, calcOriginStateHash(), trie.RootCommitment(vs.TrieAccess()))
	})
}

func TestStateWithDB(t *testing.T) {
	t.Run("save state", func(t *testing.T) {
		chainID := testmisc.RandChainID()
		store := mapdb.NewMapDB()
		vs, err := CreateOriginState(store, chainID)
		require.NoError(t, err)
		vs.Commit()
		cc := trie.RootCommitment(vs.TrieAccess())
		err = vs.Save()
		require.NoError(t, err)
		cs := trie.RootCommitment(vs.TrieAccess())
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
		su := NewStateUpdateWithBlockLogValues(1, currentTime, testmisc.RandVectorCommitment())
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

		require.EqualValues(t, trie.RootCommitment(vs1.TrieAccess()), trie.RootCommitment(vs2.TrieAccess()))
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

		require.EqualValues(t, trie.RootCommitment(vs1.TrieAccess()), trie.RootCommitment(vs2.TrieAccess()))
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
		su := NewStateUpdateWithBlockLogValues(1, time1, testmisc.RandVectorCommitment())
		su.Mutations().Set("key", []byte("value"))
		block1, err := newBlock(su.Mutations())
		require.NoError(t, err)

		err = vsOrig.ApplyBlock(block1)
		require.NoError(t, err)
		require.EqualValues(t, 1, vsOrig.BlockIndex())
		require.True(t, time1.Equal(vsOrig.Timestamp()))

		time2 := time.Now()
		su = NewStateUpdateWithBlockLogValues(2, time2, vsOrig.PreviousStateCommitment())
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

		require.EqualValues(t, trie.RootCommitment(vsOrig.TrieAccess()), trie.RootCommitment(vsLoaded.TrieAccess()))
		require.EqualValues(t, vsOrig.BlockIndex(), vsLoaded.BlockIndex())
		require.EqualValues(t, vsOrig.Timestamp(), vsLoaded.Timestamp())
		require.EqualValues(t, 2, vsLoaded.BlockIndex())

		time3 := time.Now()
		su = NewStateUpdateWithBlockLogValues(3, time3, vsLoaded.PreviousStateCommitment())
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

		require.EqualValues(t, trie.RootCommitment(vsOrig.TrieAccess()), trie.RootCommitment(vsLoaded.TrieAccess()))
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
		su := NewStateUpdateWithBlockLogValues(1, currentTime, testmisc.RandVectorCommitment())
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

func genRnd4() []string {
	str := "0123456789abcdef"
	ret := make([]string, 0, len(str)*len(str)*len(str))
	for i := range str {
		for j := range str {
			for k := range str {
				for l := range str {
					s := string([]byte{str[i], str[j], str[k], str[l]})
					s = s + s + s + s
					r1 := rand.Intn(len(s))
					r2 := rand.Intn(len(s))
					if r2 < r1 {
						r1, r2 = r2, r1
					}
					ret = append(ret, s[r1:r2])
				}
			}
		}
	}
	if len(ret) > math.MaxUint16 {
		ret = ret[:math.MaxUint16]
	}
	return ret
}

func genDifferent() []string {
	orig := genRnd4()
	// filter different
	unique := make(map[string]bool)
	for _, s := range orig {
		unique[s] = true
	}
	ret := make([]string, 0)
	for s := range unique {
		ret = append(ret, s)
	}
	return ret
}

func genRndBlocks(start, num int) []Block {
	strs := genDifferent()
	blocks := make([]Block, num)
	millis := rand.Int63()
	const numMutations = 20
	for blkNum := range blocks {
		var buf [32]byte
		copy(buf[:], fmt.Sprintf("kuku %d", blkNum))
		vc, _ := CommitmentModel.VectorCommitmentFromBytes(buf[:])
		upd := NewStateUpdateWithBlockLogValues(uint32(blkNum+start), time.UnixMilli(millis+int64(blkNum+100)), vc)
		for i := 0; i < numMutations; i++ {
			s := "1111" + strs[rand.Intn(len(strs))]
			if rand.Intn(1000) < 100 {
				upd.Mutations().Del(kv.Key(s))
			} else {
				upd.Mutations().Set(kv.Key(s), []byte(s))
			}
		}
		blocks[blkNum], _ = newBlock(upd.Mutations())
	}
	return blocks
}

func TestRnd(t *testing.T) {
	chainID := iscp.RandomChainID()

	const numBlocks = 100
	const numRepeat = 100
	cs := make([]trie.VCommitment, 0)
	blocks := genRndBlocks(2, numBlocks)
	//for bn, blk := range blocks {
	//	t.Logf("--------- #%d\nDELS: %v", bn,
	//		blk.(*blockImpl).stateUpdate.mutations.Dels)
	//}

	// blocks := genBlocks(2, numBlocks)
	t.Logf("num blocks: %d", len(blocks))
	upd1 := NewStateUpdateWithBlockLogValues(1, time.UnixMilli(0), testmisc.RandVectorCommitment())
	var exists bool
	store := make([]kvstore.KVStore, numRepeat)
	rndCommits := make([][]bool, numRepeat)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range rndCommits {
		rndCommits[i] = make([]bool, numBlocks)
		for j := range rndCommits[i] {
			rndCommits[i][j] = rng.Intn(1000) < 100
		}
	}
	// var badBlock int
	// var badKey kv.Key
	var round int
	for round = 0; round < numRepeat; round++ {
		t.Logf("------------------ round: %d", round)
		store[round] = mapdb.NewMapDB()
		vs, err := CreateOriginState(store[round], chainID)
		require.NoError(t, err)
		vs.ApplyStateUpdate(upd1)
		vs.Commit()
		c1 := trie.RootCommitment(vs.TrieAccess())
		err = vs.Save()
		require.NoError(t, err)
		c2 := trie.RootCommitment(vs.TrieAccess())
		require.True(t, trie.EqualCommitments(c1, c2))
		for bn, b := range blocks {
			require.EqualValues(t, vs.BlockIndex()+1, b.BlockIndex())
			err = vs.ApplyBlock(b)
			require.NoError(t, err)
			if rndCommits[round][bn] {
				// t.Logf("           commit at block: #%d", bn)
				err = vs.Save()
				require.NoError(t, err)
				c1 := trie.RootCommitment(vs.TrieAccess())
				vs, exists, err = LoadSolidState(store[round], chainID)
				require.NoError(t, err)
				require.True(t, exists)
				c2 := trie.RootCommitment(vs.TrieAccess())

				diff := vs.ReconcileTrie()
				if len(diff) > 0 {
					t.Logf("============== reconcile failed: %v", diff)
				}

				require.True(t, trie.EqualCommitments(c1, c2))
			}
		}
		vs.Commit()
		c1 = trie.RootCommitment(vs.TrieAccess())
		err = vs.Save()
		require.NoError(t, err)
		c2 = trie.RootCommitment(vs.TrieAccess())
		require.True(t, trie.EqualCommitments(c1, c2))

		vstmp, exists, err := LoadSolidState(store[round], chainID)
		require.NoError(t, err)
		require.True(t, exists)
		diff := vstmp.ReconcileTrie()
		if len(diff) > 0 {
			t.Logf("============== reconcile failed: %v", diff)
		}

		cs = append(cs, trie.RootCommitment(vs.TrieAccess()))
		if round > 0 {
			require.True(t, trie.EqualCommitments(cs[round-1], cs[round]))
		}
		//if round > 0 && !cs[round-1].Equal(cs[round]) {
		//	t.Logf("cs[%d] = %s != cs[%d] = %s", round-1, cs[round-1], round, cs[round])
		//	kvs0 := kv.NewHiveKVStoreReader(store[round-1])
		//	kvs1 := kv.NewHiveKVStoreReader(store[round])
		//	dkeys := kv.DumpKeySet(kv.GetDiffKeyValues(kvs0, kvs1))
		//	t.Logf("len store #%d = %d != len store #%d = %d", round-1, kv.NumKeys(kvs0), round, kv.NumKeys(kvs1))
		//
		//	t.Logf("DIFF key/values (len = %d:\n%s", len(dkeys), strings.Join(dkeys, "    \n"))
		//	diffKeys := kv.GetDiffKeys(kvs0, kvs1)
		//
		//	for k := range diffKeys {
		//		rawKey := k[1:]
		//		t.Logf("=============DIFF key: '%s'", rawKey)
		//		for bn, block := range blocks {
		//			//t.Logf("=========              block #%d", bn)
		//			mut := block.(*blockImpl).stateUpdate.mutations
		//			for x := range mut.Sets {
		//				if strings.Contains(string(x), string(rawKey)) {
		//					t.Logf("                SET in block #%d. Orig: '%s'", bn, x)
		//					badBlock = bn
		//					badKey = x
		//				}
		//			}
		//			for x := range mut.Dels {
		//				if strings.Contains(string(x), string(rawKey)) {
		//					t.Logf("                DEL in block #%d. Orig: '%s'", bn, x)
		//					badBlock = bn
		//					badKey = x
		//				}
		//			}
		//		}
		//	}
		//	t.Logf("======================= bad key '%s' in block #%d", badKey, badBlock)
		//	vs0, _, _ := LoadSolidState(store[round-1], chainID)
		//	badKeys0 := vs0.ReconcileTrie()
		//	if len(badKeys0) > 0 {
		//		t.Logf("vs0 bad keys: %v", badKeys0)
		//	}
		//	pg := vs0.ProofGeneric([]byte(badKey))
		//	t.Logf(">>>>>>>>>> vs[%d] key '%s' ending '%s', path: %d", round-1, badKey, pg.Ending, len(pg.Path))
		//	vs1, _, _ := LoadSolidState(store[round], chainID)
		//	badKeys1 := vs1.ReconcileTrie()
		//	if len(badKeys0) > 0 {
		//		t.Logf("vs1 bad keys: %v", badKeys1)
		//	}
		//	pg = vs1.ProofGeneric([]byte(badKey))
		//	t.Logf(">>>>>>>>>> vs[%d] key '%s' ending '%s', path: %d", round, badKey, pg.Ending, len(pg.Path))
		//
		//	t.Logf(">>>>>>>>> vs[%d] C = %s", round-1, vs0.RootCommitment())
		//	t.Logf(">>>>>>>>> vs[%d] C = %s", round, vs1.RootCommitment())
		//	t.FailNow()
		//}
	}
}

func TestStateBasic(t *testing.T) {
	chainID := iscp.ChainIDFromAliasID(tpkg.RandAliasAddress().AliasID())
	vs1, err := CreateOriginState(mapdb.NewMapDB(), &chainID)
	require.NoError(t, err)
	h1 := trie.RootCommitment(vs1.TrieAccess())
	require.True(t, trie.EqualCommitments(OriginStateCommitment(), h1))

	vs2 := vs1.Copy()
	h2 := trie.RootCommitment(vs2.TrieAccess())
	require.EqualValues(t, h1, h2)

	vs1.KVStore().Set(kv.Key(coreutil.StatePrefixBlockIndex), codec.EncodeUint64(1))
	vs1.KVStore().Set("num", codec.EncodeInt64(int64(123)))
	vs1.KVStore().Set("kuku", codec.EncodeString("A"))
	vs1.KVStore().Set("mumu", codec.EncodeString("B"))

	vs2.KVStore().Set(kv.Key(coreutil.StatePrefixBlockIndex), codec.EncodeUint64(1))
	vs2.KVStore().Set("mumu", codec.EncodeString("B"))
	vs2.KVStore().Set("kuku", codec.EncodeString("A"))
	vs2.KVStore().Set("num", codec.EncodeInt64(int64(123)))

	require.EqualValues(t, trie.RootCommitment(vs1.TrieAccess()), trie.RootCommitment(vs2.TrieAccess()))

	vs3 := vs1.Copy()
	vs4 := vs2.Copy()

	require.EqualValues(t, trie.RootCommitment(vs3.TrieAccess()), trie.RootCommitment(vs4.TrieAccess()))
}

func TestStateReader(t *testing.T) {
	t.Run("state not found", func(t *testing.T) {
		store := mapdb.NewMapDB()
		chainID := iscp.RandomChainID([]byte("1"))
		os, err := CreateOriginState(store, chainID)
		require.NoError(t, err)
		err = os.Save()
		require.NoError(t, err)
		c1 := trie.RootCommitment(os.TrieAccess())

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

	h1 := trie.RootCommitment(vsOpt.TrieAccess())
	require.True(t, trie.EqualCommitments(OriginStateCommitment(), h1))
	require.EqualValues(t, 0, vsOpt.BlockIndex())

	glb.InvalidateSolidIndex()
	require.PanicsWithValue(t, coreutil.ErrorStateInvalidated, func() {
		_ = trie.RootCommitment(vsOpt.TrieAccess())
	})
	require.PanicsWithValue(t, coreutil.ErrorStateInvalidated, func() {
		_ = vsOpt.BlockIndex()
	})
	require.PanicsWithValue(t, coreutil.ErrorStateInvalidated, func() {
		_, _ = vsOpt.ExtractBlock()
	})
	require.PanicsWithValue(t, coreutil.ErrorStateInvalidated, func() {
		_ = vsOpt.PreviousStateCommitment()
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

	hash := trie.RootCommitment(vs.TrieAccess())
	hashOpt := trie.RootCommitment(vsOpt.TrieAccess())
	require.EqualValues(t, hash, hashOpt)

	hashPrev := hash
	upd := NewStateUpdateWithBlockLogValues(vsOpt.BlockIndex()+1, vsOpt.Timestamp().Add(1*time.Second), trie.RootCommitment(vsOpt.TrieAccess()))
	vsOpt.ApplyStateUpdate(upd)
	hash = trie.RootCommitment(vs.TrieAccess())
	hashOpt = trie.RootCommitment(vsOpt.TrieAccess())
	require.EqualValues(t, hash, hashOpt)
	require.EqualValues(t, hashPrev, hashOpt)
}
