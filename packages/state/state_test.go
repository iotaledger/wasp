// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package state_test

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/origin"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/trie"
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

func initializedStore(db kvstore.KVStore) state.Store {
	st := state.NewStore(db)
	origin.InitChain(st, nil, 0)
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

	validateBlock0(state.NewStore(db).BlockByTrieRoot(block0.TrieRoot()))
	validateBlock0(state.NewStore(db).LatestBlock())

	require.EqualValues(t, 0, cs.LatestBlockIndex())
}

func TestOriginBlockDeterminism(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		deposit := rapid.Uint64().Draw(t, "deposit")
		db := mapdb.NewMapDB()
		st := state.NewStore(db)
		blockA := origin.InitChain(st, nil, deposit)
		blockB := origin.InitChain(st, nil, deposit)
		require.Equal(t, blockA.L1Commitment(), blockB.L1Commitment())
		db2 := mapdb.NewMapDB()
		st2 := state.NewStore(db2)
		blockC := origin.InitChain(st2, nil, deposit)
		require.Equal(t, blockA.L1Commitment(), blockC.L1Commitment())
	})
}

func Test1Block(t *testing.T) {
	db := mapdb.NewMapDB()
	cs := mustChainStore{initializedStore(db)}

	block1 := func() state.Block {
		d := cs.NewStateDraft(time.Now(), cs.LatestBlock().L1Commitment())
		d.Set("a", []byte{1})

		require.EqualValues(t, []byte{1}, d.Get("a"))

		return cs.Commit(d)
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
		block := cs.Commit(d)
		err := cs.SetLatest(block.TrieRoot())
		require.NoError(t, err)
	}

	// alt branch
	block := cs.BlockByIndex(5)
	for i := 6; i < 15; i++ {
		d := cs.NewStateDraft(time.Now(), block.L1Commitment())
		d.Set("k", []byte("b"))
		block = cs.Commit(d)
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
		d.Set("k", []byte(fmt.Sprintf("a%d", i)))
		block := cs.Commit(d)
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
		newValue := []byte(fmt.Sprintf("a%d", i))
		d1 := cs.NewStateDraft(now, latestCommitment)
		d1.Set(keyChanged, newValue)
		block1 := cs.Commit(d1)
		d2 := cs.NewStateDraft(now, latestCommitment)
		d2.Set(keyChanged, newValue)
		block2 := cs.Commit(d2)
		require.Equal(t, block1.L1Commitment(), block2.L1Commitment())
		err := cs.SetLatest(block1.TrieRoot())
		require.NoError(t, err)
	}
}

type randomState struct {
	t   *testing.T
	rnd *rand.Rand
	db  kvstore.KVStore
	cs  mustChainStore
}

func newRandomState(t *testing.T) *randomState {
	db := mapdb.NewMapDB()
	return &randomState{
		t:   t,
		rnd: rand.New(rand.NewSource(0)),
		db:  db,
		cs:  mustChainStore{initializedStore(db)},
	}
}

const rsKeyAlphabet = "ab"

// to avoid collisions with core contracts
var rsKeyPrefix = kv.Key(isc.Hn("randomState").Bytes())

func (r *randomState) randomKey() kv.Key {
	n := r.rnd.Intn(10) + 1
	b := make([]byte, n)
	for i := 0; i < n; i++ {
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

func (r *randomState) commitNewBlock(latestBlock state.Block, timestamp time.Time) state.Block {
	d := r.cs.NewStateDraft(timestamp, latestBlock.L1Commitment())
	for j := 0; j < 50; j++ {
		d.Set(r.randomKey(), r.randomValue())
	}
	for j := 0; j < 10; j++ {
		d.Del(r.randomKey())
	}
	block := r.cs.Commit(d)
	err := r.cs.SetLatest(block.TrieRoot())
	require.NoError(r.t, err)
	return block
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

		for i := 1; i <= 20; i++ {
			block := r.commitNewBlock(r.cs.LatestBlock(), time.Unix(int64(i), 0))

			index := block.StateIndex()
			t.Logf("committed block %d", index)

			if keepLatest > 0 && index >= uint32(keepLatest) {
				p := index - uint32(keepLatest)
				trieRoot := r.cs.BlockByIndex(p).TrieRoot()
				stats, err := r.cs.Prune(trieRoot)
				require.NoError(t, err)

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
		block := r.commitNewBlock(r.cs.LatestBlock(), time.Unix(n, 0))
		n++
		trieRoots = append(trieRoots, block.TrieRoot())
	}

	r.rnd.Shuffle(len(trieRoots), func(i, j int) {
		trieRoots[i], trieRoots[j] = trieRoots[j], trieRoots[i]
	})

	for len(trieRoots) > 3 {
		// prune 2 random trie roots
		for i := 0; i < 2; i++ {
			trieRoot := trieRoots[0]
			stats, err := r.cs.Prune(trieRoot)
			require.NoError(t, err)
			t.Logf("pruned trie root %x: %+v", trieRoot, stats)
			trieRoots = trieRoots[1:]

			for _, trieRoot := range trieRoots {
				r.cs.checkTrie(trieRoot)
			}
		}

		// commit a new block based off a random trie root
		trieRoot := trieRoots[0]
		block := r.commitNewBlock(r.cs.BlockByTrieRoot(trieRoot), time.Unix(n, 0))
		n++
		trieRoots = append(trieRoots, block.TrieRoot())
		t.Logf("committed block: %d", len(trieRoots))
	}
}

func TestSnapshot(t *testing.T) {
	snapshot := mapdb.NewMapDB()

	trieRoot, blockHash := func() (trie.Hash, state.BlockHash) {
		db := mapdb.NewMapDB()
		cs := mustChainStore{initializedStore(db)}
		for i := byte(1); i <= 10; i++ {
			d := cs.NewStateDraft(time.Now(), cs.LatestBlock().L1Commitment())
			d.Set(kv.Key(fmt.Sprintf("k%d", i)), []byte("v"))
			d.Set("k", []byte{i})
			block := cs.Commit(d)
			err := cs.SetLatest(block.TrieRoot())
			require.NoError(t, err)
		}
		block := cs.LatestBlock()
		err := cs.TakeSnapshot(block.TrieRoot(), snapshot)
		require.NoError(t, err)
		return block.TrieRoot(), block.Hash()
	}()

	db := mapdb.NewMapDB()
	cs := mustChainStore{state.NewStore(db)}
	err := cs.RestoreSnapshot(trieRoot, snapshot)
	require.NoError(t, err)

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
}
