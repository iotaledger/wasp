package testmisc

import (
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/trie"
	"github.com/iotaledger/wasp/packages/kv/trie_merkle"
)

func RandChainID() *iscp.ChainID {
	ret := iscp.ChainIDFromAliasID(tpkg.RandAliasAddress().AliasID())
	return &ret
}

func RandVectorCommitment() trie.VectorCommitment {
	h := hashing.RandomHash(nil)
	ret, err := trie_merkle.NewVectorCommitmentFromBytes(h[:])
	if err != nil {
		panic(err)
	}
	return ret
}
