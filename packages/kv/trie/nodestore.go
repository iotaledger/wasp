package trie

import (
	"bytes"
	"reflect"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// CommitmentModel abstracts 256+ Trie logic from the commitment logic/cryptography
type CommitmentModel interface {
	NewVectorCommitment() VCommitment
	NewTerminalCommitment() TCommitment
	VectorCommitmentFromBytes([]byte) (VCommitment, error)
	CommitToNode(*Node) VCommitment
	CommitToData([]byte) TCommitment
}

// NodeStore is an interface to nodeStore to the trie as a set of nodeStore represented as key/value pairs
// Two implementations:
// - nodeStore is a direct, non-cached nodeStore to key/value storage
// - Trie implement a cached nodeStore
type NodeStore interface {
	GetNode(key kv.Key) (*Node, bool)
	Model() CommitmentModel
}

func NewNodeStore(store kv.KVMustReader, model CommitmentModel) *nodeStore { //nolint:revive
	return &nodeStore{
		model: model,
		store: store,
	}
}

// nodeStore direct nodeStore to trie
type nodeStore struct {
	model CommitmentModel
	store kv.KVMustReader
}

func (tr *nodeStore) GetNode(key kv.Key) (*Node, bool) {
	nodeBin := tr.store.MustGet(key)
	if nodeBin == nil {
		return nil, false
	}
	node, err := NodeFromBytes(tr.model, nodeBin)
	assert(err == nil, err)
	return node, true
}

func (tr *nodeStore) Model() CommitmentModel {
	return tr.model
}

func RootCommitment(tr NodeStore) VCommitment {
	n, ok := tr.GetNode("")
	if !ok {
		return nil
	}
	return tr.Model().CommitToNode(n)
}

func EqualCommitments(c1, c2 CommitmentBase) bool {
	if c1 == c2 {
		return true
	}
	// TODO better suggestion ? The problem: type(nil) != nil
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
