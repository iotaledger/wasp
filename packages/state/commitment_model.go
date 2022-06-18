package state

import (
	"bytes"
	"errors"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/trie.go/hive_adaptor"
	"github.com/iotaledger/trie.go/models/trie_blake2b"
	"github.com/iotaledger/trie.go/models/trie_blake2b/trie_blake2b_verify"
	"github.com/iotaledger/trie.go/trie"
	"github.com/iotaledger/wasp/packages/database/dbkeys"
)

// singleton of CommitmentModel. In ISC we use the following trie parameters:
// - hexary trie
// - 20 byte blake2b hashing
// - valueSizeOptimizationThreshold = 64
// We do not optimize key commitments in ISC state
var (
	model                = trie_blake2b.New(trie.PathArity16, trie_blake2b.HashSize160, 64)
	vectorCommitmentSize = len(model.NewVectorCommitment().Bytes())
)

func trieKVStore(db kvstore.KVStore) trie.KVStore {
	return hive_adaptor.NewHiveKVStoreAdaptor(db, []byte{dbkeys.ObjectTypeTrie})
}

func valueKVStore(db kvstore.KVStore) trie.KVStore {
	return hive_adaptor.NewHiveKVStoreAdaptor(db, []byte{dbkeys.ObjectTypeState})
}

func NewTrie(db kvstore.KVStore) *trie.Trie {
	return trie.New(model, trieKVStore(db), valueKVStore(db))
}

func NewTrieReader(trieKV, valueKV trie.KVReader) *trie.TrieReader {
	return trie.NewTrieReader(model, trieKV, valueKV)
}

func RootCommitment(tr trie.NodeStore) trie.VCommitment {
	return trie.RootCommitment(tr)
}

func EqualCommitments(c1, c2 trie.Serializable) bool {
	return model.EqualCommitments(c1, c2)
}

func VCommitmentFromBytes(data []byte) (trie.VCommitment, error) {
	if len(data) != vectorCommitmentSize {
		return nil, errors.New("wrong data size")
	}
	ret := model.NewVectorCommitment()
	if err := ret.Read(bytes.NewReader(data)); err != nil {
		return nil, err
	}
	return ret, nil
}

func GetMerkleProof(key []byte, tr trie.NodeStore) *trie_blake2b.Proof {
	return model.Proof(key, tr)
}

func ValidateMerkleProof(proof *trie_blake2b.Proof, root trie.VCommitment, value ...[]byte) error {
	if len(value) == 0 {
		return trie_blake2b_verify.Validate(proof, root.Bytes())
	}
	return trie_blake2b_verify.ValidateWithValue(proof, root.Bytes(), value[0])
}
