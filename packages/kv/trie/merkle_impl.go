package trie

import (
	"encoding/hex"
	"golang.org/x/crypto/blake2b"
	"io"
)

// implements commitment scheme based on blake2b hashing

type hashCommitment [32]byte

// MerkleTrieSetup implements 256+ trie based on Merkle tree, i.e. on hashing with blake2b
var MerkleTrieSetup = &TrieSetup{
	NewTerminalCommitment: newTerminalCommitment,
	NewVectorCommitment:   newVectorCommitment,
	CommitToChildren:      commitToChildren,
	CommitToData:          commitToData,
	UpdateCommitment:      updateVectorCommitment,
	UpdateNodeCommitment:  updateNodeCommitment,
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
	empty := true
	for i := range n.children {
		pos := 32 * int(i)
		n.children[i].Write(sliceWriter(buf[pos : pos+32]))
		empty = false
	}
	if n.terminal != nil {
		n.terminal.Write(sliceWriter(buf[256*32:]))
		empty = false
	}
	if empty {
		return nil
	}
	ret := hashCommitment(blake2b.Sum256(buf[:]))
	return &ret
}

func commitToData(data []byte) TerminalCommitment {
	if len(data) == 0 {
		// empty slice -> no data (deleted)
		return nil
	}
	ret := hashCommitment{}
	if len(data) <= 32 {
		copy(ret[:], data)
	} else {
		ret = blake2b.Sum256(data)
	}
	return &ret
}

func updateVectorCommitment(prev *VectorCommitment, delta VectorCommitment) {
	*prev = delta
}

func updateNodeCommitment(n *Node) VectorCommitment {
	n.terminal = n.newTerminal
	for i, child := range n.modifiedChildren {
		c := updateNodeCommitment(child)
		if c != nil {
			n.children[i] = c
		} else {
			// deletion
			delete(n.children, i)
		}
	}
	n.modifiedChildren = make(map[byte]*Node)
	ret := commitToChildren(n)
	assert((ret == nil) == n.IsEmpty(), "assert: (ret==nil) == n.IsEmpty()")
	return ret
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

func (s *hashCommitment) Equal(another commitment) bool {
	if s == nil && another == nil {
		return true
	}
	if s == nil || another == nil {
		return false
	}
	a, ok := another.(*hashCommitment)
	if !ok {
		return false
	}
	return *s == *a
}
