// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package state

import (
	"golang.org/x/crypto/blake2b"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/isc/coreutil"
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/kv/buffered"
	"github.com/iotaledger/wasp/v2/packages/kv/codec"
	"github.com/iotaledger/wasp/v2/packages/kvstore/mapdb"
	"github.com/iotaledger/wasp/v2/packages/trie"
)

type block struct {
	trieRoot             trie.Hash           `bcs:"export"`
	mutations            *buffered.Mutations `bcs:"export"`
	previousL1Commitment *L1Commitment       `bcs:"export,optional"`
}

var _ Block = &block{}

// BlockHashCreate needs to be public for bcs to function
type BlockHashCreate struct {
	Sets kv.Items
	Dels []kv.Key
}

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
	// TODO: The hash should be calculated by bcs.MustMarshal(b.Mutations()), but it currently does not get properly sorted
	// TODO: Once the bcs ordering is fixed, remove the Create struct + MustMarshal here and Marshal(b.Mutations()) instead.
	bb := BlockHashCreate{
		Sets: b.mutations.Clone().SetsSorted(),
		Dels: b.mutations.Clone().DelsSorted(),
	}

	data := bcs.MustMarshal[BlockHashCreate](&bb)

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
	return codec.MustDecode[uint32](b.MutationsReader().Get(kv.Key(coreutil.StatePrefixBlockIndex)))
}

func (b *block) TrieRoot() trie.Hash {
	return b.trieRoot
}
