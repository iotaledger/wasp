// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package state_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
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
