package trie

import (
	"bytes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"reflect"
)

// CommitmentModel abstracts 256+ Trie logic from the commitment logic/cryptography
type CommitmentModel interface {
	NewVectorCommitment() VCommitment
	NewTerminalCommitment() TCommitment
	VectorCommitmentFromBytes([]byte) (VCommitment, error)
	CommitToNode(*Node) VCommitment
	CommitToData([]byte) TCommitment
}

// Access is an interface for read only access to the trie
type Access interface {
	GetNode(key kv.Key) (*Node, bool)
	Model() CommitmentModel
}

func NewTrieAccess(store kv.KVMustReader, model CommitmentModel) *accessTrie {
	return &accessTrie{
		model: model,
		store: store,
	}
}

type accessTrie struct {
	model CommitmentModel
	store kv.KVMustReader
}

func (tr *accessTrie) GetNode(key kv.Key) (*Node, bool) {
	nodeBin := tr.store.MustGet(key)
	if nodeBin == nil {
		return nil, false
	}
	node, err := NodeFromBytes(tr.model, nodeBin)
	assert(err == nil, err)
	return node, true
}

func (tr *accessTrie) Model() CommitmentModel {
	return tr.model
}

func RootCommitment(tr Access) VCommitment {
	n, ok := tr.GetNode("")
	if !ok {
		return nil
	}
	return tr.Model().CommitToNode(n)
}

// GetProofGeneric returns generic proof path. Contains references trie node cache.
// Should be immediately converted into the specific proof model independent of the trie
// Normally only called by the model
func GetProofGeneric(tr Access, key []byte) *ProofGeneric {
	if len(key) == 0 {
		key = []byte{}
	}
	p, _, ending := proofPath(tr, key)
	return &ProofGeneric{
		Key:    key,
		Path:   p,
		Ending: ending,
	}
}

// returns
// - path of keys which leads to 'path'
// - common prefix between the last key and the fragment
func proofPath(trieAccess Access, path []byte) ([][]byte, []byte, ProofEndingCode) {
	node, ok := trieAccess.GetNode("")
	if !ok {
		return nil, nil, 0
	}

	proof := make([][]byte, 0)
	var key []byte

	for {
		// it means we continue the branch of commitment
		proof = append(proof, key)
		assert(len(key) <= len(path), "len(key) <= len(path)")
		if bytes.Equal(path[len(key):], node.PathFragment) {
			return proof, nil, EndingTerminal
		}
		prefix := commonPrefix(path[len(key):], node.PathFragment)

		if len(prefix) < len(node.PathFragment) {
			return proof, prefix, EndingSplit
		}
		assert(len(prefix) == len(node.PathFragment), "len(prefix)==len(node.PathFragment)")
		childIndexPosition := len(key) + len(prefix)
		assert(childIndexPosition < len(path), "childIndexPosition<len(path)")

		childKey := node.ChildKey(kv.Key(key), path[childIndexPosition])

		node, ok = trieAccess.GetNode(childKey)
		if !ok {
			// if there are no commitment to the child at the position, it means trie must be extended at this point
			return proof, prefix, EndingExtend
		}
		key = []byte(childKey)
	}
}

func EqualCommitments(c1, c2 CommitmentBase) bool {
	if c1 == c2 {
		return true
	}
	c1Nil := c1 == nil || (reflect.ValueOf(c1).Kind() == reflect.Ptr && reflect.ValueOf(c1).IsNil())
	c2Nil := c2 == nil || (reflect.ValueOf(c2).Kind() == reflect.Ptr && reflect.ValueOf(c2).IsNil())
	if c1Nil && c2Nil {
		return true
	}
	if c1Nil || c2Nil {
		return false
	}
	return bytes.Equal(c1.Bytes(), c2.Bytes())
}

// ComputeCommitment computes commitment to arbitrary key/value iterator
func ComputeCommitment(model CommitmentModel, store kv.KVIterator) (VCommitment, error) {
	emptyStore := dict.New()
	tr := New(model, emptyStore)

	err := store.Iterate("", func(key kv.Key, value []byte) bool {
		tr.Update([]byte(key), value)
		return true
	})
	if err != nil {
		return nil, err
	}
	tr.Commit()
	return RootCommitment(tr), nil
}
