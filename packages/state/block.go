// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package state

import (
	"math"
	"math/rand"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/trie"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

type block struct {
	trieRoot             trie.Hash           `bcs:""`
	mutations            *buffered.Mutations `bcs:""`
	previousL1Commitment *L1Commitment       `bcs:"optional"`
}

var _ Block = &block{}

func NewBlock() Block {
	return new(block)
}

func BlockFromBytes(data []byte) (Block, error) {
	return bcs.Unmarshal[*block](data)
}

func (b *block) Bytes() []byte {
	return bcs.MustMarshal(b)
}

func (b *block) Equals(other Block) bool {
	return b.Hash().Equals(other.Hash())
}

// Hash calculates a hash from the mutations and previousL1Commitment
func (b *block) Hash() (ret BlockHash) {
	data := bcs.MustMarshal(b.mutations)
	if b.previousL1Commitment != nil {
		data = append(data, bcs.MustMarshal(b.previousL1Commitment)...)
	}
	hash := blake2b.Sum256(data)
	copy(ret[:], hash[:])
	return ret
}

func (b *block) L1Commitment() *L1Commitment {
	return newL1Commitment(b.TrieRoot(), b.Hash())
}

func (b *block) Mutations() *buffered.Mutations {
	return b.mutations
}

func (b *block) MutationsReader() kv.KVStoreReader {
	return buffered.NewBufferedKVStoreForMutations(
		kv.NewHiveKVStoreReader(mapdb.NewMapDB()),
		b.mutations,
	)
}

func (b *block) PreviousL1Commitment() *L1Commitment {
	return b.previousL1Commitment
}

func (b *block) StateIndex() uint32 {
	return codec.Uint32.MustDecode(b.MutationsReader().Get(kv.Key(coreutil.StatePrefixBlockIndex)))
}

func (b *block) TrieRoot() trie.Hash {
	return b.trieRoot
}

// test only function
func RandomBlock() Block {
	store := NewStoreWithUniqueWriteMutex(mapdb.NewMapDB())
	draft := store.NewOriginStateDraft()
	for i := 0; i < 3; i++ {
		draft.Set(kv.Key([]byte{byte(rand.Intn(math.MaxInt8))}), []byte{byte(rand.Intn(math.MaxInt8))})
	}

	return store.Commit(draft)
}
