package state

import (
	"testing"

	"github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/kv/buffered"
	"github.com/iotaledger/wasp/v2/packages/testutil/testval"
	"github.com/iotaledger/wasp/v2/packages/trie"
)

func TestBlockCodec(t *testing.T) {
	muts := buffered.NewMutations()
	muts.Set(kv.Key(testval.TestBytes(10, 1)), testval.TestBytes(17, 2))
	muts.Set(kv.Key(testval.TestBytes(5, 3)), testval.TestBytes(34, 4))
	muts.Del(kv.Key(testval.TestBytes(9, 5)), true)

	var b Block = &block{
		trieRoot:  trie.RandomHash(),
		mutations: muts,
		previousL1Commitment: &L1Commitment{
			trieRoot:  trie.RandomHash(),
			blockHash: RandomBlockHash(),
		},
	}
	bcs.TestCodec(t, b)

	b = &block{
		trieRoot:  trie.TestHash,
		mutations: muts,
		previousL1Commitment: &L1Commitment{
			trieRoot:  trie.Hash(testval.TestBytes(trie.HashSizeBytes)),
			blockHash: TestBlockHash,
		},
	}
	bcs.TestCodecAndHash(t, b, "8533a81db889")
}
