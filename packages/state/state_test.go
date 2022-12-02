// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package state

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/core/kvstore/mapdb"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/trie.go/models/trie_blake2b/trie_blake2b_verify"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

func TestOriginBlock(t *testing.T) {
	db := mapdb.NewMapDB()
	emptyChainID := isc.ChainID{}

	cs := InitChainStore(db)

	validateBlock0 := func(block0 Block) {
		require.True(t, block0.PreviousTrieRoot() == nil)
		require.False(t, block0.TrieRoot() == nil)
		require.EqualValues(t, map[kv.Key][]byte{
			KeyChainID:                             emptyChainID.Bytes(),
			kv.Key(coreutil.StatePrefixBlockIndex): codec.EncodeUint32(0),
			kv.Key(coreutil.StatePrefixTimestamp):  codec.EncodeTime(time.Unix(0, 0)),
		}, block0.Mutations().Sets)
		require.Empty(t, block0.Mutations().Dels)
	}

	block0 := cs.BlockByIndex(0)
	validateBlock0(block0)
	state := cs.StateByTrieRoot(block0.TrieRoot())
	require.True(t, state.ChainID().Equals(&emptyChainID))
	require.EqualValues(t, 0, state.BlockIndex())
	require.True(t, state.Timestamp().IsZero())

	validateBlock0(NewStore(db).BlockByTrieRoot(block0.TrieRoot()))
	validateBlock0(NewStore(db).BlockByIndex(0))

	require.EqualValues(t, 0, cs.LatestBlockIndex())
}

func TestApprovingOID(t *testing.T) {
	db := mapdb.NewMapDB()
	cs := InitChainStore(db)
	oid := tpkg.RandUTXOInput()
	block0 := cs.BlockByIndex(0)
	cs.SetApprovingOutputID(block0.TrieRoot(), oid)
	require.True(t, cs.BlockByTrieRoot(block0.TrieRoot()).ApprovingOutputID().Equals(oid))
}

func Test1Block(t *testing.T) {
	db := mapdb.NewMapDB()
	cs := InitChainStore(db)

	chainID := isc.RandomChainID()

	block1 := func() Block {
		d := cs.NewStateDraft(time.Now(), cs.LatestBlock().L1Commitment())
		d.Set("a", []byte{1})
		d.Set(KeyChainID, chainID.Bytes())

		require.EqualValues(t, []byte{1}, d.MustGet("a"))
		require.True(t, d.ChainID().Equals(chainID))

		return cs.Commit(d)
	}()
	cs.SetLatest(block1.TrieRoot())
	require.EqualValues(t, 1, cs.LatestBlockIndex())

	require.EqualValues(t, 0, cs.StateByIndex(0).BlockIndex())
	require.EqualValues(t, 1, cs.StateByIndex(1).BlockIndex())
	require.EqualValues(t, []byte{1}, cs.BlockByIndex(1).Mutations().Sets["a"])

	oid := tpkg.RandUTXOInput()
	cs.SetApprovingOutputID(block1.TrieRoot(), oid)
	require.True(t, oid.Equals(cs.BlockByIndex(1).ApprovingOutputID()))

	require.EqualValues(t, []byte{1}, cs.StateByIndex(1).MustGet("a"))
	require.True(t, cs.StateByIndex(1).ChainID().Equals(chainID))
}

func TestReorg(t *testing.T) {
	db := mapdb.NewMapDB()
	cs := InitChainStore(db)

	// main branch
	for i := 1; i < 10; i++ {
		d := cs.NewStateDraft(time.Now(), cs.LatestBlock().L1Commitment())
		d.Set("k", []byte("a"))
		block := cs.Commit(d)
		cs.SetLatest(block.TrieRoot())
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
	cs.SetLatest(block.TrieRoot())
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

func TestProof(t *testing.T) {
	db := mapdb.NewMapDB()
	cs := InitChainStore(db)

	for _, k := range [][]byte{
		[]byte(KeyChainID),
		[]byte(coreutil.StatePrefixTimestamp),
		[]byte(coreutil.StatePrefixBlockIndex),
	} {
		t.Run(fmt.Sprintf("%x", k), func(t *testing.T) {
			v := cs.LatestState().MustGet(kv.Key(k))
			require.NotEmpty(t, v)

			proof := cs.LatestState().GetMerkleProof([]byte(k))
			require.False(t, trie_blake2b_verify.IsProofOfAbsence(proof))
			err := ValidateMerkleProof(proof, cs.LatestBlock().TrieRoot(), v)
			require.NoError(t, err)
		})
	}
}
