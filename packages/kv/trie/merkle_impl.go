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
	for i := range n.children {
		pos := 32 * int(i)
		n.children[i].Write(sliceWriter(buf[pos : pos+32]))
	}
	if n.terminalCommitment != nil {
		n.terminalCommitment.Write(sliceWriter(buf[256*32:]))
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

func updateVectorCommitment(prev *VectorCommitment, delta VectorCommitment) {
	*prev = delta
}

func updateNodeCommitment(n *Node) VectorCommitment {
	n.terminalCommitment = n.newTerminal
	for i, child := range n.modifiedChildren {
		n.children[i] = updateNodeCommitment(child)
	}
	n.modifiedChildren = make(map[uint8]*Node)
	return commitToChildren(n)
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
