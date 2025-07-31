// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package state_test

import (
	"bytes"
	"fmt"
	"hash/crc32"
	"math/rand"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/packages/chaindb"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/isc/coreutil"
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/kvstore"
	"github.com/iotaledger/wasp/v2/packages/kvstore/mapdb"
	"github.com/iotaledger/wasp/v2/packages/origin"
	"github.com/iotaledger/wasp/v2/packages/parameters/parameterstest"
	"github.com/iotaledger/wasp/v2/packages/state"
	"github.com/iotaledger/wasp/v2/packages/state/statetest"
	"github.com/iotaledger/wasp/v2/packages/trie"
	"github.com/iotaledger/wasp/v2/packages/vm/core/migrations/allmigrations"
)

type mustChainStore struct {
	state.Store
}

func (m mustChainStore) BlockByIndex(i uint32) state.Block {
	latest, err := m.Store.LatestBlock()
	if err != nil {
		panic(err)
	}
	if i > latest.StateIndex() {
		panic(fmt.Sprintf("invalid index %d (latest is %d)", i, latest.StateIndex()))
	}
	block := latest
	for block.StateIndex() > i {
		block, err = m.Store.BlockByTrieRoot(block.PreviousL1Commitment().TrieRoot())
		if err != nil {
			panic(err)
		}
	}
	return block
}

func (m mustChainStore) StateByIndex(i uint32) state.State {
	block := m.BlockByIndex(i)
	state, err := m.Store.StateByTrieRoot(block.TrieRoot())
	if err != nil {
		panic(err)
	}
	return state
}

func (m mustChainStore) LatestState() state.State {
	r, err := m.Store.LatestState()
	if err != nil {
		panic(err)
	}
	return r
}

func (m mustChainStore) StateByTrieRoot(root trie.Hash) state.State {
	r, err := m.Store.StateByTrieRoot(root)
	if err != nil {
		panic(err)
	}
	return r
}

func (m mustChainStore) BlockByTrieRoot(root trie.Hash) state.Block {
	r, err := m.Store.BlockByTrieRoot(root)
	if err != nil {
		panic(err)
	}
	return r
}

func (m mustChainStore) LatestBlock() state.Block {
	r, err := m.Store.LatestBlock()
	if err != nil {
		panic(err)
	}
	return r
}

func (m mustChainStore) LatestBlockIndex() uint32 {
	r, err := m.Store.LatestBlockIndex()
	if err != nil {
		panic(err)
	}
	return r
}

func (m mustChainStore) NewStateDraft(timestamp time.Time, prevL1Commitment *state.L1Commitment) state.StateDraft {
	r, err := m.Store.NewStateDraft(timestamp, prevL1Commitment)
	if err != nil {
		panic(err)
	}
	return r
}

// check that the trie can access all its nodes
func (m mustChainStore) checkTrie(trieRoot trie.Hash) {
	m.StateByTrieRoot(trieRoot).Iterate("", func(k kv.Key, v []byte) bool {
		return true
	})
}

var initArgs = origin.DefaultInitParams(isc.NewAddressAgentID(cryptolib.NewEmptyAddress())).Encode()

func initializedStore(db kvstore.KVStore) state.Store {
	st := statetest.NewStoreWithUniqueWriteMutex(db)
	origin.InitChain(allmigrations.LatestSchemaVersion, st, initArgs, iotago.ObjectID{}, 0, parameterstest.L1Mock)
	return st
}

func TestOriginBlock(t *testing.T) {
	db := mapdb.NewMapDB()

	cs := mustChainStore{initializedStore(db)}

	validateBlock0 := func(block0 state.Block, err error) {
		require.NoError(t, err)
		require.True(t, block0.PreviousL1Commitment() == nil)
		require.Empty(t, block0.Mutations().Dels)
	}

	block0 := cs.BlockByIndex(0)
	validateBlock0(block0, nil)
	s := cs.StateByTrieRoot(block0.TrieRoot())
	require.EqualValues(t, 0, s.BlockIndex())
	require.True(t, s.Timestamp().IsZero())

	validateBlock0(statetest.NewStoreWithUniqueWriteMutex(db).BlockByTrieRoot(block0.TrieRoot()))
	validateBlock0(statetest.NewStoreWithUniqueWriteMutex(db).LatestBlock())

	require.EqualValues(t, 0, cs.LatestBlockIndex())
}

func TestOriginBlockDeterminism(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		deposit := coin.Value(rapid.Uint64().Draw(t, "deposit"))
		db := mapdb.NewMapDB()
		st := statetest.NewStoreWithUniqueWriteMutex(db)
		require.True(t, st.IsEmpty())
		blockA, _ := origin.InitChain(allmigrations.LatestSchemaVersion, st, initArgs, iotago.ObjectID{}, deposit, parameterstest.L1Mock)
		blockB, _ := origin.InitChain(allmigrations.LatestSchemaVersion, st, initArgs, iotago.ObjectID{}, deposit, parameterstest.L1Mock)
		require.False(t, st.IsEmpty())
		require.Equal(t, blockA.L1Commitment(), blockB.L1Commitment())
		db2 := mapdb.NewMapDB()
		st2 := statetest.NewStoreWithUniqueWriteMutex(db2)
		require.True(t, st2.IsEmpty())
		blockC, _ := origin.InitChain(allmigrations.LatestSchemaVersion, st2, initArgs, iotago.ObjectID{}, deposit, parameterstest.L1Mock)
		require.False(t, st2.IsEmpty())
		require.Equal(t, blockA.L1Commitment(), blockC.L1Commitment())
	})
}

func Test1Block(t *testing.T) {
	db := mapdb.NewMapDB()
	cs := mustChainStore{initializedStore(db)}
	require.False(t, cs.IsEmpty())

	block1 := func() state.Block {
		d := cs.NewStateDraft(time.Now(), cs.LatestBlock().L1Commitment())
		d.Set("a", []byte{1})

		require.EqualValues(t, []byte{1}, d.Get("a"))
		block, _, _ := lo.Must3(cs.Commit(d))
		require.False(t, cs.IsEmpty())

		return block
	}()
	err := cs.SetLatest(block1.TrieRoot())
	require.NoError(t, err)
	require.EqualValues(t, 1, cs.LatestBlockIndex())

	require.EqualValues(t, 0, cs.StateByIndex(0).BlockIndex())
	require.EqualValues(t, 1, cs.StateByIndex(1).BlockIndex())
	require.EqualValues(t, []byte{1}, cs.BlockByIndex(1).Mutations().Sets["a"])

	require.EqualValues(t, []byte{1}, cs.StateByIndex(1).Get("a"))
}

func TestReorg(t *testing.T) {
	db := mapdb.NewMapDB()
	cs := mustChainStore{initializedStore(db)}

	// main branch
	for i := 1; i < 10; i++ {
		d := cs.NewStateDraft(time.Now(), cs.LatestBlock().L1Commitment())
		d.Set("k", []byte("a"))
		block, _, _ := lo.Must3(cs.Commit(d))
		err := cs.SetLatest(block.TrieRoot())
		require.NoError(t, err)
	}

	// alt branch
	block := cs.BlockByIndex(5)
	for i := 6; i < 15; i++ {
		d := cs.NewStateDraft(time.Now(), block.L1Commitment())
		d.Set("k", []byte("b"))
		block, _, _ = lo.Must3(cs.Commit(d))
	}

	// no reorg yet
	require.EqualValues(t, 9, cs.LatestBlockIndex())
	for i := uint32(1); i <= cs.LatestBlockIndex(); i++ {
		require.EqualValues(t, i, cs.StateByIndex(i).BlockIndex())
		require.EqualValues(t, []byte("a"), cs.StateByIndex(i).Get("k"))
	}

	// reorg
	err := cs.SetLatest(block.TrieRoot())
	require.NoError(t, err)
	require.EqualValues(t, 14, cs.LatestBlockIndex())
	for i := uint32(1); i <= cs.LatestBlockIndex(); i++ {
		t.Log(i)
		require.EqualValues(t, i, cs.StateByIndex(i).BlockIndex())
		if i <= 5 {
			require.EqualValues(t, []byte("a"), cs.StateByIndex(i).Get("k"))
		} else {
			require.EqualValues(t, []byte("b"), cs.StateByIndex(i).Get("k"))
		}
	}
}

func TestReplay(t *testing.T) {
	db := mapdb.NewMapDB()
	cs := mustChainStore{initializedStore(db)}
	for i := 1; i < 10; i++ {
		d := cs.NewStateDraft(time.Now(), cs.LatestBlock().L1Commitment())
		d.Set("k", fmt.Appendf(nil, "a%d", i))
		block, _, _ := lo.Must3(cs.Commit(d))
		err := cs.SetLatest(block.TrieRoot())
		require.NoError(t, err)
	}

	// create a clone of the store by replaying all the blocks
	db2 := mapdb.NewMapDB()
	cs2 := mustChainStore{initializedStore(db2)}
	for i := 1; i < 10; i++ {
		block := cs.BlockByIndex(uint32(i))

		d, err := cs2.NewEmptyStateDraft(block.PreviousL1Commitment())
		require.NoError(t, err)
		block.Mutations().ApplyTo(d)
		cs2.Commit(d)
	}
	err := cs2.SetLatest(cs.LatestBlock().TrieRoot())
	require.NoError(t, err)
}

func TestEqualStates(t *testing.T) {
	db1 := mapdb.NewMapDB()
	cs1 := mustChainStore{initializedStore(db1)}
	time1 := time.Now()
	draft1 := cs1.NewStateDraft(time1, origin.L1Commitment(allmigrations.LatestSchemaVersion, initArgs, iotago.ObjectID{}, 0, parameterstest.L1Mock))
	draft1.Set("a", []byte("variable a"))
	draft1.Set("b", []byte("variable b"))
	block1, _, _ := lo.Must3(cs1.Commit(draft1))
	time2 := time.Now()
	draft2 := cs1.NewStateDraft(time2, block1.L1Commitment())
	draft2.Set("b", []byte("another value of b"))
	draft2.Set("c", []byte("new variable c"))
	block2, _, _ := lo.Must3(cs1.Commit(draft2))
	time3 := time.Now()
	draft3 := cs1.NewStateDraft(time3, block2.L1Commitment())
	draft3.Del("a")
	draft3.Set("d", []byte("newest variable d"))
	block3, _, _ := lo.Must3(cs1.Commit(draft3))
	state1 := cs1.StateByTrieRoot(block3.TrieRoot())

	db2 := mapdb.NewMapDB()
	cs2 := mustChainStore{initializedStore(db2)}
	draft1 = cs2.NewStateDraft(time1, origin.L1Commitment(allmigrations.LatestSchemaVersion, initArgs, iotago.ObjectID{}, 0, parameterstest.L1Mock))
	draft1.Set("b", []byte("variable b"))
	draft1.Set("a", []byte("variable a"))
	block1, _, _ = lo.Must3(cs2.Commit(draft1))
	draft2 = cs2.NewStateDraft(time2, block1.L1Commitment())
	draft2.Set("c", []byte("new variable c"))
	draft2.Set("b", []byte("another value of b"))
	block2, _, _ = lo.Must3(cs2.Commit(draft2))
	draft3 = cs2.NewStateDraft(time3, block2.L1Commitment())
	draft3.Set("d", []byte("newest variable d"))
	draft3.Del("a")
	block3, _, _ = lo.Must3(cs2.Commit(draft3))
	state2 := cs2.StateByTrieRoot(block3.TrieRoot())

	require.True(t, state1.Equals(state2))
	require.True(t, state2.Equals(state1))
	require.True(t, state1.TrieRoot().Equals(state2.TrieRoot()))
	require.Equal(t, state1.BlockIndex(), state2.BlockIndex())
	require.Equal(t, state1.Timestamp(), state2.Timestamp())
	require.True(t, state1.PreviousL1Commitment().Equals(state2.PreviousL1Commitment()))
	commonState := getCommonState(state1, state2)
	for _, entry := range commonState {
		require.True(t, bytes.Equal(entry.value1, entry.value2))
	}
}

type commonEntry struct {
	value1 []byte
	value2 []byte
}

func getCommonState(state1, state2 state.State) map[kv.Key]*commonEntry {
	result := make(map[kv.Key]*commonEntry)
	iterateFun := func(iterState state.State, setValueFun func(*commonEntry, []byte)) {
		iterState.Iterate(kv.EmptyPrefix, func(key kv.Key, value []byte) bool {
			entry, ok := result[key]
			if !ok {
				entry = &commonEntry{}
				result[key] = entry
			}
			setValueFun(entry, value)
			return true
		})
	}
	iterateFun(state1, func(entry *commonEntry, value []byte) { entry.value1 = value })
	iterateFun(state2, func(entry *commonEntry, value []byte) { entry.value2 = value })
	return result
}

func TestDiffStatesValues(t *testing.T) {
	db1 := mapdb.NewMapDB()
	cs1 := mustChainStore{initializedStore(db1)}
	time1 := time.Now()
	draft1 := cs1.NewStateDraft(time1, origin.L1Commitment(allmigrations.LatestSchemaVersion, initArgs, iotago.ObjectID{}, 0, parameterstest.L1Mock))
	draft1.Set("a", []byte("variable a"))
	block1, _, _ := lo.Must3(cs1.Commit(draft1))
	state1 := cs1.StateByTrieRoot(block1.TrieRoot())

	db2 := mapdb.NewMapDB()
	cs2 := mustChainStore{initializedStore(db2)}
	draft1 = cs2.NewStateDraft(time1, origin.L1Commitment(allmigrations.LatestSchemaVersion, initArgs, iotago.ObjectID{}, 0, parameterstest.L1Mock))
	draft1.Set("a", []byte("other value of a"))
	block1, _, _ = lo.Must3(cs2.Commit(draft1))
	state2 := cs2.StateByTrieRoot(block1.TrieRoot())

	require.False(t, state1.Equals(state2))
	require.False(t, state2.Equals(state1))
}

func TestDiffStatesBlockIndex(t *testing.T) {
	db1 := mapdb.NewMapDB()
	cs1 := mustChainStore{initializedStore(db1)}
	time1 := time.Now()
	draft1 := cs1.NewStateDraft(time1, origin.L1Commitment(allmigrations.LatestSchemaVersion, initArgs, iotago.ObjectID{}, 0, parameterstest.L1Mock))
	draft1.Set("a", []byte("variable a"))
	block1, _, _ := lo.Must3(cs1.Commit(draft1))
	time2 := time.Now()
	draft2 := cs1.NewStateDraft(time2, block1.L1Commitment())
	draft1.Set("b", []byte("variable b"))
	block2, _, _ := lo.Must3(cs1.Commit(draft2))
	state1 := cs1.StateByTrieRoot(block2.TrieRoot())

	db2 := mapdb.NewMapDB()
	cs2 := mustChainStore{initializedStore(db2)}
	draft1 = cs2.NewStateDraft(time1, origin.L1Commitment(allmigrations.LatestSchemaVersion, initArgs, iotago.ObjectID{}, 0, parameterstest.L1Mock))
	draft1.Set("a", []byte("variable a"))
	draft1.Set("b", []byte("variable b"))
	block1, _, _ = lo.Must3(cs2.Commit(draft1))
	state2 := cs2.StateByTrieRoot(block1.TrieRoot())

	require.Equal(t, uint32(2), state1.BlockIndex())
	require.Equal(t, uint32(1), state2.BlockIndex())
	require.False(t, state1.Equals(state2))
	require.False(t, state2.Equals(state1))
}

func TestDiffStatesTimestamp(t *testing.T) {
	db1 := mapdb.NewMapDB()
	cs1 := mustChainStore{initializedStore(db1)}
	draft1 := cs1.NewStateDraft(time.Now(), origin.L1Commitment(allmigrations.LatestSchemaVersion, initArgs, iotago.ObjectID{}, 0, parameterstest.L1Mock))
	draft1.Set("a", []byte("variable a"))
	block1, _, _ := lo.Must3(cs1.Commit(draft1))
	state1 := cs1.StateByTrieRoot(block1.TrieRoot())

	db2 := mapdb.NewMapDB()
	cs2 := mustChainStore{initializedStore(db2)}
	draft1 = cs2.NewStateDraft(time.Now(), origin.L1Commitment(allmigrations.LatestSchemaVersion, initArgs, iotago.ObjectID{}, 0, parameterstest.L1Mock))
	draft1.Set("a", []byte("variable a"))
	block1, _, _ = lo.Must3(cs2.Commit(draft1))
	state2 := cs2.StateByTrieRoot(block1.TrieRoot())

	require.NotEqual(t, state1.Timestamp(), state2.Timestamp())
	require.False(t, state1.Equals(state2))
	require.False(t, state2.Equals(state1))
}

func TestProof(t *testing.T) {
	db := mapdb.NewMapDB()
	cs := mustChainStore{initializedStore(db)}

	for _, k := range [][]byte{
		[]byte(coreutil.StatePrefixTimestamp),
		[]byte(coreutil.StatePrefixBlockIndex),
	} {
		t.Run(fmt.Sprintf("%x", k), func(t *testing.T) {
			v := cs.LatestState().Get(kv.Key(k))
			require.NotEmpty(t, v)

			proof := cs.LatestState().GetMerkleProof(k)
			require.False(t, proof.IsProofOfAbsence())
			err := proof.ValidateValue(cs.LatestBlock().TrieRoot(), v)
			require.NoError(t, err)
		})
	}
}

func TestDoubleCommit(t *testing.T) {
	db := mapdb.NewMapDB()
	cs := mustChainStore{initializedStore(db)}
	keyChanged := kv.Key("k")
	for i := 1; i < 10; i++ {
		now := time.Now()
		latestCommitment := cs.LatestBlock().L1Commitment()
		newValue := fmt.Appendf(nil, "a%d", i)
		d1 := cs.NewStateDraft(now, latestCommitment)
		d1.Set(keyChanged, newValue)
		block1, _, _ := lo.Must3(cs.Commit(d1))
		d2 := cs.NewStateDraft(now, latestCommitment)
		d2.Set(keyChanged, newValue)
		block2, _, _ := lo.Must3(cs.Commit(d2))
		require.Equal(t, block1.L1Commitment(), block2.L1Commitment())
		err := cs.SetLatest(block1.TrieRoot())
		require.NoError(t, err)
	}
}

type randomState struct {
	t   testing.TB
	rnd *rand.Rand
	db  kvstore.KVStore
	cs  mustChainStore
}

func newRandomStateWithDB(t testing.TB, db kvstore.KVStore) *randomState {
	return &randomState{
		t:   t,
		rnd: rand.New(rand.NewSource(0)),
		db:  db,
		cs:  mustChainStore{initializedStore(db)},
	}
}

func newRandomState(t *testing.T) *randomState {
	return newRandomStateWithDB(t, mapdb.NewMapDB())
}

const rsKeyAlphabet = "ab"

// to avoid collisions with core contracts
var rsKeyPrefix = kv.Key(isc.Hn("randomState").Bytes())

func (r *randomState) randomKey() kv.Key {
	n := r.rnd.Intn(10) + 1
	b := make([]byte, n)
	for i := range n {
		b[i] = rsKeyAlphabet[r.rnd.Intn(len(rsKeyAlphabet))]
	}
	return rsKeyPrefix + kv.Key(b)
}

func (r *randomState) randomValue() []byte {
	// half of the values will be stored in the trie nodes,
	// and the other half in the value store
	n := r.rnd.Intn(128) + 1
	b := make([]byte, n)
	_, err := r.rnd.Read(b)
	require.NoError(r.t, err)
	return b
}

func (r *randomState) commitNewBlock(latestBlock state.Block, timestamp time.Time) (state.Block, bool, *trie.CommitStats) {
	d := r.cs.NewStateDraft(timestamp, latestBlock.L1Commitment())
	for range 50 {
		d.Set(r.randomKey(), r.randomValue())
	}
	for range 10 {
		d.Del(r.randomKey())
	}
	block, refcountsEnabled, stats, err := r.cs.Commit(d)
	require.NoError(r.t, err)
	err = r.cs.SetLatest(block.TrieRoot())
	require.NoError(r.t, err)
	return block, refcountsEnabled, stats
}

func dbSize(db kvstore.KVStore) int {
	size := 0
	err := db.Iterate(kvstore.EmptyPrefix, func(k []byte, v []byte) bool {
		size += len(k) + len(v)
		return true
	})
	if err != nil {
		panic(err)
	}
	return size
}

func TestPruning(t *testing.T) {
	run := func(keepLatest int) int {
		var sizes []int

		r := newRandomState(t)
		t.Logf("committed block %d, %s", 0, r.cs.LatestBlock().TrieRoot())

		for i := 1; i <= 20; i++ {
			block, _, stats := r.commitNewBlock(r.cs.LatestBlock(), time.Unix(int64(i), 0))

			index := block.StateIndex()
			t.Logf("committed block %d, %+v, %s", index, stats, block.TrieRoot())

			if keepLatest > 0 && index >= uint32(keepLatest) {
				p := index - uint32(keepLatest)
				trieRoot := r.cs.BlockByIndex(p).TrieRoot()
				stats, err := r.cs.Prune(trieRoot)
				require.NoError(t, err)
				lpbIndex, err := r.cs.LargestPrunedBlockIndex()
				require.NoError(t, err)
				require.Equal(t, p, lpbIndex)

				t.Logf("pruned block %d: %+v %s", p, stats, trieRoot)
				{
					_, err := r.cs.Store.StateByTrieRoot(trieRoot)
					require.ErrorContains(t, err, "does not exist")
				}
				{
					_, err := r.cs.Store.BlockByTrieRoot(trieRoot)
					require.ErrorContains(t, err, "not found")
				}
				require.False(t, r.cs.HasTrieRoot(trieRoot))
			} else {
				_, err := r.cs.LargestPrunedBlockIndex()
				require.Error(t, err)
			}

			sizes = append(sizes, dbSize(r.db))

			r.cs.checkTrie(r.cs.LatestBlock().TrieRoot())
		}

		t.Log(sizes)
		return sizes[len(sizes)-1]
	}

	dbSizeWithoutPruning := run(0)
	dbSizeWithPruning := run(10)

	require.Less(t, dbSizeWithPruning, dbSizeWithoutPruning)
}

func TestPruning2(t *testing.T) {
	r := newRandomState(t)

	trieRoots := []trie.Hash{r.cs.LatestBlock().TrieRoot()}

	var n int64 = 1

	for i := 1; i <= 20; i++ {
		block, _, stats := r.commitNewBlock(r.cs.LatestBlock(), time.Unix(n, 0))
		t.Logf("committed block %d, %+v", block.StateIndex(), stats)
		n++
		trieRoots = append(trieRoots, block.TrieRoot())
	}

	r.rnd.Shuffle(len(trieRoots), func(i, j int) {
		trieRoots[i], trieRoots[j] = trieRoots[j], trieRoots[i]
	})

	lpbIndexExpected := uint32(0)
	_, err := r.cs.LargestPrunedBlockIndex()
	require.Error(t, err)
	for len(trieRoots) > 3 {
		// prune 2 random trie roots
		for range 2 {
			trieRoot := trieRoots[0]
			block := r.cs.BlockByTrieRoot(trieRoot)
			stats, err := r.cs.Prune(trieRoot)
			require.NoError(t, err)
			if block.StateIndex() > lpbIndexExpected {
				lpbIndexExpected = block.StateIndex()
			}
			lpbIndex, err := r.cs.LargestPrunedBlockIndex()
			require.NoError(t, err)
			require.Equal(t, lpbIndexExpected, lpbIndex)
			t.Logf("pruned trie root %x: %+v", trieRoot, stats)
			trieRoots = trieRoots[1:]

			for _, trieRoot := range trieRoots {
				r.cs.checkTrie(trieRoot)
			}
		}

		// commit a new block based off a random trie root
		trieRoot := trieRoots[0]
		block, _, stats := r.commitNewBlock(r.cs.BlockByTrieRoot(trieRoot), time.Unix(n, 0))
		t.Logf("committed block %d, %+v", block.StateIndex(), stats)
		n++
		trieRoots = append(trieRoots, block.TrieRoot())
		t.Logf("committed block: %d", len(trieRoots))
	}
}

func makeRandomDB(t *testing.T, nBlocks int) (mustChainStore, kvstore.KVStore) {
	db := mapdb.NewMapDB()
	cs := mustChainStore{initializedStore(db)}
	require.False(t, cs.IsEmpty())
	for i := 1; i <= nBlocks; i++ {
		d := cs.NewStateDraft(time.Now(), cs.LatestBlock().L1Commitment())
		d.Set(kv.Key(fmt.Sprintf("k%d", i)), []byte("v"))
		d.Set("k", []byte{byte(i)})
		if i == 1 {
			d.Set("x", []byte(strings.Repeat("v", 70)))
		}
		block, _, _ := lo.Must3(cs.Commit(d))
		require.False(t, cs.IsEmpty())
		err := cs.SetLatest(block.TrieRoot())
		require.NoError(t, err)
	}
	return cs, db
}

func makeRandomDBSnapshot(t *testing.T, nBlocks int) (trie.Hash, state.BlockHash, *bytes.Buffer) {
	cs, _ := makeRandomDB(t, nBlocks)
	block := cs.LatestBlock()
	snapshot := new(bytes.Buffer)
	err := cs.TakeSnapshot(block.TrieRoot(), snapshot)
	require.NoError(t, err)
	return block.TrieRoot(), block.Hash(), snapshot
}

func TestSnapshot(t *testing.T) {
	trieRoot, blockHash, snapshot := makeRandomDBSnapshot(t, 10)

	db := mapdb.NewMapDB()
	cs := mustChainStore{statetest.NewStoreWithUniqueWriteMutex(db)}
	require.True(t, cs.IsEmpty())
	err := cs.RestoreSnapshot(trieRoot, bytes.NewReader(snapshot.Bytes()), true)
	require.NoError(t, err)
	cs.SetLatest(trieRoot)
	require.False(t, cs.IsEmpty())

	block := cs.LatestBlock()
	require.EqualValues(t, 10, block.StateIndex())
	require.EqualValues(t, blockHash, block.Hash())

	_, err = cs.Store.BlockByTrieRoot(block.PreviousL1Commitment().TrieRoot())
	require.ErrorContains(t, err, "not found")

	state := cs.LatestState()
	for i := byte(1); i <= 10; i++ {
		require.EqualValues(t, []byte("v"), state.Get(kv.Key(fmt.Sprintf("k%d", i))))
	}
	require.EqualValues(t, []byte{10}, state.Get("k"))
	require.EqualValues(t, []byte(strings.Repeat("v", 70)), state.Get("x"))
}

func TestRestoreSnapshotEmptyDB(t *testing.T) {
	trieRoot, _, snapshot := makeRandomDBSnapshot(t, 10)

	// restore the snapshot on empty DB
	db := mapdb.NewMapDB()
	cs := mustChainStore{statetest.NewStoreWithUniqueWriteMutex(db)}
	err := cs.RestoreSnapshot(trieRoot, bytes.NewReader(snapshot.Bytes()), true)
	require.NoError(t, err)

	// at this point the DB contains a single trie root with all refcounts = 1
	// let's prune it and assert that the DB is left (almost) empty: as pruning
	// adds largest pruned block index into the store, it still remains there
	// even if all the other information is deleted. See addLargestPrunedBlockIndex
	// for details.
	_, err = cs.Prune(trieRoot)
	require.NoError(t, err)
	expectedMap := addLargestPrunedBlockIndex(map[string][]byte{}, 10)
	require.EqualValues(t, expectedMap, toMap(db))
}

func TestRestoreSnapshotNonEmptyDB(t *testing.T) {
	trieRoot, _, snapshot := makeRandomDBSnapshot(t, 10)

	cs, db := makeRandomDB(t, 10)
	dbCopy := toMap(db)

	// restore the snapshot, then prune it -- the DB should be left unchanged,
	// except largest pruned block index, which is added after pruning. See
	// addLargestPrunedBlockIndex for details.
	err := cs.RestoreSnapshot(trieRoot, bytes.NewReader(snapshot.Bytes()), true)
	require.NoError(t, err)
	_, err = cs.Prune(trieRoot)
	require.NoError(t, err)

	dbCopy2 := toMap(db)

	require.EqualValues(t, addLargestPrunedBlockIndex(dbCopy, 10), dbCopy2)
}

func TestPrunedSnapshot(t *testing.T) {
	r := newRandomState(t)
	for i := 1; i <= 20; i++ {
		block, _, stats := r.commitNewBlock(r.cs.LatestBlock(), time.Now())
		require.False(t, r.cs.IsEmpty())
		index := block.StateIndex()
		t.Logf("committed block %d, %+v", index, stats)
	}
	_, err := r.cs.LargestPrunedBlockIndex()
	require.Error(t, err)

	for i := 0; i <= 10; i++ {
		block := r.cs.BlockByIndex(uint32(i))
		var stats trie.PruneStats
		stats, err = r.cs.Prune(block.TrieRoot())
		require.NoError(t, err)
		var lpbIndex uint32
		lpbIndex, err = r.cs.LargestPrunedBlockIndex()
		require.NoError(t, err)
		require.Equal(t, uint32(i), lpbIndex)
		require.False(t, r.cs.IsEmpty())
		t.Logf("pruned trie block index %v: %+v", i, stats)
	}

	blockToSnapshot := r.cs.LatestBlock()
	snapshot := new(bytes.Buffer)
	err = r.cs.TakeSnapshot(blockToSnapshot.TrieRoot(), snapshot)
	require.NoError(t, err)
	require.False(t, r.cs.IsEmpty())
	t.Logf("snapshotted block index %v", blockToSnapshot.StateIndex())

	db := mapdb.NewMapDB()
	cs := mustChainStore{statetest.NewStoreWithUniqueWriteMutex(db)}
	require.True(t, cs.IsEmpty())
	err = cs.RestoreSnapshot(blockToSnapshot.TrieRoot(), bytes.NewReader(snapshot.Bytes()), true)
	require.NoError(t, err)
	_, err = cs.LargestPrunedBlockIndex()
	require.Error(t, err)
	require.False(t, cs.IsEmpty())
}

func TestRefcountsToggle(t *testing.T) {
	// calculate hash from trie data (excluding refcounts)
	calculateHash := func(store mustChainStore, latestTrieRoot trie.Hash) uint32 {
		h := crc32.NewIEEE()
		block := store.BlockByTrieRoot(latestTrieRoot)
		for {
			store.StateByTrieRoot(block.TrieRoot()).Iterate("", func(key kv.Key, value []byte) bool {
				h.Write([]byte(key))
				h.Write(value)
				return true
			})
			prev := block.PreviousL1Commitment()
			if prev == nil {
				break
			}
			block = store.BlockByTrieRoot(block.PreviousL1Commitment().TrieRoot())
		}
		return h.Sum32()
	}

	r := newRandomState(t)
	for i := 1; i <= 20; i++ {
		block, _, stats := r.commitNewBlock(r.cs.LatestBlock(), time.Now())
		require.False(t, r.cs.IsEmpty())
		index := block.StateIndex()
		t.Logf("committed block %d, %+v", index, stats)
	}
	require.True(t, r.cs.IsRefcountsEnabled())
	hash1 := calculateHash(r.cs, r.cs.LatestBlock().TrieRoot())

	// refcounts can be disabled anytime
	newStore, err := state.NewStoreWithMetrics(r.db, false, new(sync.Mutex), nil)
	require.NoError(t, err)
	storeWithoutRefcounts := mustChainStore{Store: newStore}
	require.False(t, r.cs.IsRefcountsEnabled())
	// trie can be iterated and produces the same hash
	hash2 := calculateHash(storeWithoutRefcounts, r.cs.LatestBlock().TrieRoot())
	require.Equal(t, hash1, hash2)

	// can commit a new block
	r.commitNewBlock(r.cs.LatestBlock(), time.Now())
	// trie can be iterated from the new block
	hash3 := calculateHash(storeWithoutRefcounts, r.cs.LatestBlock().TrieRoot())
	require.NotEqual(t, hash1, hash3)
	// trie can be iterated from the previous block and produces the same hash
	hash4 := calculateHash(storeWithoutRefcounts, r.cs.LatestBlock().PreviousL1Commitment().TrieRoot())
	require.Equal(t, hash1, hash4)

	// attempting to prune produces an error
	_, err = storeWithoutRefcounts.Prune(storeWithoutRefcounts.LatestBlock().TrieRoot())
	require.ErrorContains(t, err, "refcounts disabled")

	// once disabled, refcounts cannot be enabled again
	_, err = state.NewStoreWithMetrics(r.db, true, new(sync.Mutex), nil)
	require.ErrorContains(t, err, "non-empty store")
}

func toMap(store kvstore.KVStore) map[string][]byte {
	m := make(map[string][]byte)
	store.Iterate(kvstore.EmptyPrefix, func(k, v []byte) bool {
		m[string(k)] = v
		return true
	})
	return m
}

// Just for testing; works for small indexes (0-127). Key of added entry is
// `[]byte{3}` (see keyLargestPrunedBlockIndex function) converted to string.
// Value of added entry is state index put in four bytes little endian format.
// If state index is not larger than 127, its value fits in the least significant
// byte and other three bytes are 0.
func addLargestPrunedBlockIndex(db map[string][]byte, indexOfLeastSignificantByte byte) map[string][]byte {
	db[string([]byte{chaindb.PrefixLargestPrunedBlockIndex})] = []byte{indexOfLeastSignificantByte, 0, 0, 0}

	// refcounts enabled flag
	db[string([]byte{chaindb.PrefixTrie, 4})] = []byte{1}
	return db
}
