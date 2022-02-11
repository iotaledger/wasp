package trie

import (
	"encoding/hex"
	"golang.org/x/crypto/blake2b"
	"io"
)

// implements commitment scheme based on blake2b hashing

type hashCommitment [32]byte

var MerkleTrieSetup = &TrieSetup{
	NewTerminalCommitment: newTerminalCommitment,
	NewVectorCommitment:   newVectorCommitment,
	CommitToChildren:      commitToChildren,
	CommitToData:          commitToData,
	UpdateKey:             updateKey,
}

func newTerminalCommitment() TerminalCommitment {
	return &hashCommitment{}
}

func newVectorCommitment() VectorCommitment {
	return &hashCommitment{}
}

type sliceWriter []byte

func (w sliceWriter) Write(p []byte) (int, error) {
	if len(p) > len(w) {
		panic("sliceWriter: data does not fit the target")
	}
	copy(w, p)
	return len(p), nil
}

func commitToChildren(n *Node) VectorCommitment {
	var buf [257 * 32]byte // 8 KB + 32 B
	pos := 0
	for i := range n.children {
		if n.children[i] != nil {
			n.children[i].Write(sliceWriter(buf[pos : pos+32]))
		}
		pos += 32
	}
	if n.terminalCommitment != nil {
		n.terminalCommitment.Write(sliceWriter(buf[pos : pos+32]))
	}
	ret := hashCommitment(blake2b.Sum256(buf[:]))
	return &ret
}

func commitToData(data []byte) TerminalCommitment {
	ret := hashCommitment{}
	if len(data) <= 32 {
		copy(ret[:], data)
	} else {
		ret = blake2b.Sum256(data)
	}
	return &ret
}

func updateKey(t *trie, path []byte, pathPosition int, updateCommitment *VectorCommitment, terminal TerminalCommitment) {
	assert(pathPosition <= len(path), "pathPosition <= len(path)")
	if len(path) == 0 {
		path = []byte{}
	}
	key := path[:pathPosition]
	node, ok := t.GetNode(key)
	if !ok {
		// node for the path[:pathPosition] does not exist
		// create a new one, put rest of the path into the fragment
		// Commit to terminal value
		var err error
		node = t.mustNewNode(key)
		assert(err == nil, err)

		node.pathFragment = path[pathPosition:]
		node.terminalCommitment = terminal
		*updateCommitment = t.setup.CommitToChildren(node)
		return
	}
	// node for the path[:pathPosition] exists
	prefix := commonPrefix(node.pathFragment, path[pathPosition:])
	assert(len(prefix) <= len(node.pathFragment), "len(prefix)<= len(node.pathFragment)")
	// the following parameters define how it goes:
	// - len(path)
	// - pathPosition
	// - len(node.pathFragment)
	// - len(prefix)
	nextPathPosition := pathPosition + len(prefix)
	assert(nextPathPosition <= len(path), "nextPathPosition <= len(path)")

	if len(prefix) == len(node.pathFragment) {
		// pathFragment is part of the path. No need for a fork, continue the path
		if nextPathPosition == len(path) {
			// reached the terminal value on this node
			node.terminalCommitment = terminal
			*updateCommitment = t.setup.CommitToChildren(node)
		} else {
			assert(nextPathPosition < len(path), "nextPathPosition < len(path)")
			// didn't reach the end of the path
			// choose direction and continue down the path of the child
			childIndex := path[nextPathPosition]

			// recursively update the rest of the path
			t.setup.UpdateKey(t, path, nextPathPosition+1, &node.children[childIndex], terminal)
			*updateCommitment = t.setup.CommitToChildren(node)
		}
		return
	}
	assert(len(prefix) < len(node.pathFragment), "len(prefix) < len(node.pathFragment)")

	// need for the fork of the pathFragment
	// continued branch is part of the fragment
	keyContinue := make([]byte, pathPosition+len(prefix)+1)
	copy(keyContinue, path)
	keyContinue[len(keyContinue)-1] = node.pathFragment[len(prefix)]

	// nodeContinue continues old path
	nodeContinue := t.mustNewNode(keyContinue)
	nodeContinue.pathFragment = node.pathFragment[len(prefix)+1:]
	nodeContinue.children = node.children
	nodeContinue.terminalCommitment = node.terminalCommitment

	// adjust the old node. It will hold 2 commitments to the forked nodes
	childIndexContinue := keyContinue[len(keyContinue)-1]
	node.pathFragment = prefix
	node.children = [256]VectorCommitment{}
	node.terminalCommitment = nil

	// previous commitment must exist
	assert(*updateCommitment != nil, "*updateCommitment != nil")
	node.children[childIndexContinue] = *updateCommitment

	if pathPosition+len(prefix) == len(path) {
		// no need for the new node
		node.terminalCommitment = terminal
	} else {
		// create the new node
		keyFork := path[:pathPosition+len(prefix)+1]
		assert(len(keyContinue) == len(keyFork), "len(keyContinue)==len(keyFork)")
		nodeFork := t.mustNewNode(keyFork)
		nodeFork.pathFragment = path[len(keyFork):]
		nodeFork.terminalCommitment = terminal
		childForkIndex := keyFork[len(keyFork)-1]
		node.children[childForkIndex] = t.setup.CommitToChildren(nodeFork)
	}
	*updateCommitment = t.setup.CommitToChildren(node)
}

func (s *hashCommitment) Read(r io.Reader) error {
	_, err := r.Read((*s)[:])
	return err
}

func (s *hashCommitment) Write(w io.Writer) {
	_, _ = w.Write((*s)[:])
}

func (s *hashCommitment) String() string {
	return hex.EncodeToString(s[:])
}
