// Copyright 2022 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package state

import (
	"io"
	"math"
	"math/rand"

	"golang.org/x/crypto/blake2b"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/trie"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type block struct {
	trieRoot             trie.Hash
	mutations            *buffered.Mutations
	previousL1Commitment *L1Commitment
}

var _ Block = &block{}

func NewBlock() Block {
	return new(block)
}

func BlockFromBytes(data []byte) (Block, error) {
	return rwutil.ReadFromBytes(data, new(block))
}

func (b *block) Bytes() []byte {
	return rwutil.WriteToBytes(b)
}

func (b *block) essenceBytes() []byte {
	ww := rwutil.NewBytesWriter()
	b.writeEssence(ww)
	return ww.Bytes()
}

func (b *block) Equals(other Block) bool {
	return b.Hash().Equals(other.Hash())
}

func (b *block) Hash() (ret BlockHash) {
	hash := blake2b.Sum256(b.essenceBytes())
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
	return codec.MustDecodeUint32(b.MutationsReader().Get(kv.Key(coreutil.StatePrefixBlockIndex)))
}

func (b *block) TrieRoot() trie.Hash {
	return b.trieRoot
}

func (b *block) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	rr.ReadN(b.trieRoot[:])
	b.readEssence(rr)
	return rr.Err
}

func (b *block) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteN(b.trieRoot[:])
	b.writeEssence(ww)
	return ww.Err
}

func (b *block) readEssence(rr *rwutil.Reader) {
	b.mutations = buffered.NewMutations()
	rr.Read(b.mutations)
	hasPrevL1Commitment := rr.ReadBool()
	if hasPrevL1Commitment {
		b.previousL1Commitment = new(L1Commitment)
		rr.Read(b.previousL1Commitment)
	}
}

func (b *block) writeEssence(ww *rwutil.Writer) {
	ww.Write(b.mutations)
	ww.WriteBool(b.previousL1Commitment != nil)
	if b.previousL1Commitment != nil {
		ww.Write(b.previousL1Commitment)
	}
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
