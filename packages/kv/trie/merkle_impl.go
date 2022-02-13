package trie

import (
	"encoding/hex"
	"golang.org/x/crypto/blake2b"
	"io"
)

// implements commitment scheme based on blake2b hashing

type hashCommitment [32]byte

// MerkleCommitments implements 256+ Trie based on Merkle tree, i.e. on hashing with blake2b
type merkleTrieSetup struct{}

var (
	MerkleCommitments                 = &merkleTrieSetup{}
	_                 CommitmentLogic = MerkleCommitments
)

func (s *merkleTrieSetup) NewTerminalCommitment() TerminalCommitment {
	return &hashCommitment{}
}

func (s *merkleTrieSetup) NewVectorCommitment() VectorCommitment {
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

func (s *merkleTrieSetup) CommitToNode(n *Node) VectorCommitment {
	var buf [258 * 32]byte // 8 KB + 32 B + 32 B
	empty := true
	for i := range n.children {
		pos := 32 * int(i)
		n.children[i].Write(sliceWriter(buf[pos : pos+32]))
		empty = false
	}
	if n.terminal != nil {
		n.terminal.Write(sliceWriter(buf[256*32 : 256*32+32]))
		empty = false
	}
	if empty {
		return nil
	}
	// committing to the pathFragment. To be able to prove absence of key
	hashData(n.pathFragment).Write(sliceWriter(buf[257*32 : 257*32+32]))
	ret := hashCommitment(blake2b.Sum256(buf[:]))
	return &ret
}

func hashData(data []byte) *hashCommitment {
	ret := hashCommitment{}
	if len(data) <= 32 {
		copy(ret[:], data)
	} else {
		ret = blake2b.Sum256(data)
	}
	return &ret
}

func (s *merkleTrieSetup) CommitToData(data []byte) TerminalCommitment {
	if len(data) == 0 {
		// empty slice -> no data (deleted)
		return nil
	}
	return hashData(data)
}

func (s *merkleTrieSetup) UpdateCommitment(prev *VectorCommitment, delta VectorCommitment) {
	*prev = delta
}

func (s *merkleTrieSetup) UpdateNodeCommitment(n *Node) VectorCommitment {
	n.terminal = n.newTerminal
	for i, child := range n.modifiedChildren {
		c := s.UpdateNodeCommitment(child)
		if c != nil {
			n.children[i] = c
		} else {
			// deletion
			delete(n.children, i)
		}
	}
	n.modifiedChildren = make(map[byte]*Node)
	ret := s.CommitToNode(n)
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
