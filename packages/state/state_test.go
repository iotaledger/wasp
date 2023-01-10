// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package state

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/core/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/trie"
)

type mustChainStore struct {
	Store
}

func (m mustChainStore) BlockByIndex(i uint32) Block {
	r, err := m.Store.BlockByIndex(i)
	if err != nil {
		panic(err)
	}
	return r
}

func (m mustChainStore) StateByIndex(i uint32) State {
	r, err := m.Store.StateByIndex(i)
	if err != nil {
		panic(err)
	}
	return r
}

func (m mustChainStore) LatestState() State {
	r, err := m.Store.LatestState()
	if err != nil {
		panic(err)
	}
	return r
}

func (m mustChainStore) StateByTrieRoot(root trie.Hash) State {
	r, err := m.Store.StateByTrieRoot(root)
	if err != nil {
		panic(err)
	}
	return r
}

func (m mustChainStore) BlockByTrieRoot(root trie.Hash) Block {
	r, err := m.Store.BlockByTrieRoot(root)
	if err != nil {
		panic(err)
	}
	return r
}

func (m mustChainStore) LatestBlock() Block {
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

func (m mustChainStore) NewStateDraft(timestamp time.Time, prevL1Commitment *L1Commitment) StateDraft {
	r, err := m.Store.NewStateDraft(timestamp, prevL1Commitment)
	if err != nil {
		panic(err)
	}
	return r
}

func TestOriginBlock(t *testing.T) {
	db := mapdb.NewMapDB()

	cs := mustChainStore{InitChainStore(db)}

	validateBlock0 := func(block0 Block, err error) {
		require.NoError(t, err)
		require.True(t, block0.PreviousL1Commitment() == nil)
		require.EqualValues(t, map[kv.Key][]byte{
			KeyChainID:                             isc.EmptyChainID().Bytes(),
			kv.Key(coreutil.StatePrefixBlockIndex): codec.EncodeUint32(0),
			kv.Key(coreutil.StatePrefixTimestamp):  codec.EncodeTime(time.Unix(0, 0)),
		}, block0.Mutations().Sets)
		require.Empty(t, block0.Mutations().Dels)
	}

	block0 := cs.BlockByIndex(0)
	validateBlock0(block0, nil)
	state := cs.StateByTrieRoot(block0.TrieRoot())
	require.True(t, state.ChainID().Empty())
	require.EqualValues(t, 0, state.BlockIndex())
	require.True(t, state.Timestamp().IsZero())

	validateBlock0(NewStore(db).BlockByTrieRoot(block0.TrieRoot()))
	validateBlock0(NewStore(db).BlockByIndex(0))

	require.EqualValues(t, 0, cs.LatestBlockIndex())
}

func Test1Block(t *testing.T) {
	db := mapdb.NewMapDB()
	cs := mustChainStore{InitChainStore(db)}

	chainID := isc.RandomChainID()

	block1 := func() Block {
		d := cs.NewStateDraft(time.Now(), cs.LatestBlock().L1Commitment())
		d.Set("a", []byte{1})
		d.Set(KeyChainID, chainID.Bytes())

		require.EqualValues(t, []byte{1}, d.MustGet("a"))
		require.True(t, d.ChainID().Equals(chainID))

		return cs.Commit(d)
	}()
	err := cs.SetLatest(block1.TrieRoot())
	require.NoError(t, err)
	require.EqualValues(t, 1, cs.LatestBlockIndex())

	require.EqualValues(t, 0, cs.StateByIndex(0).BlockIndex())
	require.EqualValues(t, 1, cs.StateByIndex(1).BlockIndex())
	require.EqualValues(t, []byte{1}, cs.BlockByIndex(1).Mutations().Sets["a"])

	require.EqualValues(t, []byte{1}, cs.StateByIndex(1).MustGet("a"))
	require.True(t, cs.StateByIndex(1).ChainID().Equals(chainID))
}

func TestReorg(t *testing.T) {
	db := mapdb.NewMapDB()
	cs := mustChainStore{InitChainStore(db)}

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
		require.EqualValues(t, []byte("a"), cs.StateByIndex(i).MustGet("k"))
	}

	// reorg
	err := cs.SetLatest(block.TrieRoot())
	require.NoError(t, err)
	require.EqualValues(t, 14, cs.LatestBlockIndex())
	for i := uint32(1); i <= cs.LatestBlockIndex(); i++ {
		t.Log(i)
		require.EqualValues(t, i, cs.StateByIndex(i).BlockIndex())
		if i <= 5 {
			require.EqualValues(t, []byte("a"), cs.StateByIndex(i).MustGet("k"))
		} else {
			require.EqualValues(t, []byte("b"), cs.StateByIndex(i).MustGet("k"))
		}
	}
}

func TestReplay(t *testing.T) {
	db := mapdb.NewMapDB()
	cs := mustChainStore{InitChainStore(db)}
	for i := 1; i < 10; i++ {
		d := cs.NewStateDraft(time.Now(), cs.LatestBlock().L1Commitment())
		d.Set("k", []byte(fmt.Sprintf("a%d", i)))
		block := cs.Commit(d)
		err := cs.SetLatest(block.TrieRoot())
		require.NoError(t, err)
	}

	// create a clone of the store by replaying all the blocks
	db2 := mapdb.NewMapDB()
	cs2 := mustChainStore{InitChainStore(db2)}
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
	cs := mustChainStore{InitChainStore(db)}

	for _, k := range [][]byte{
		[]byte(KeyChainID),
		[]byte(coreutil.StatePrefixTimestamp),
		[]byte(coreutil.StatePrefixBlockIndex),
	} {
		t.Run(fmt.Sprintf("%x", k), func(t *testing.T) {
			v := cs.LatestState().MustGet(kv.Key(k))
			require.NotEmpty(t, v)

			proof := cs.LatestState().GetMerkleProof(k)
			require.False(t, proof.IsProofOfAbsence())
			err := proof.ValidateValue(cs.LatestBlock().TrieRoot(), v)
			require.NoError(t, err)
		})
	}
}
