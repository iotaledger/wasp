package merkle_trie

import (
	"encoding/hex"
	"github.com/iotaledger/wasp/packages/kv/trie"
	"golang.org/x/crypto/blake2b"
	"io"
)

// implements commitment scheme based on blake2b hashing

type hashCommitment [32]byte

// CommitmentLogic implements 256+ Trie based on Merkle tree, i.e. on hashing with blake2b
type trieSetup struct{}

var (
	CommitmentLogic                      = &trieSetup{}
	_               trie.CommitmentLogic = CommitmentLogic
)

func (m *trieSetup) NewTerminalCommitment() trie.TerminalCommitment {
	return &hashCommitment{}
}

func (m *trieSetup) NewVectorCommitment() trie.VectorCommitment {
	return &hashCommitment{}
}

func (m *trieSetup) CommitToNode(n *trie.Node) trie.VectorCommitment {
	var hashes [258]*hashCommitment

	empty := true
	for i, c := range n.Children {
		hashes[i] = c.(*hashCommitment)
		empty = false
	}
	if n.Terminal != nil {
		hashes[256] = n.Terminal.(*hashCommitment)
		empty = false
	}
	if empty {
		return nil
	}
	hashes[257] = hashData(n.PathFragment)
	return hashHashes(&hashes)
}

func hashHashes(hashes *[258]*hashCommitment) *hashCommitment {
	var buf [258 * 32]byte // 8 KB + 32 B + 32 B
	for i, h := range hashes {
		if h == nil {
			continue
		}
		pos := 32 * int(i)
		h.Write(trie.NewSliceWriter(buf[pos : pos+32]))
	}
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

func (m *trieSetup) CommitToData(data []byte) trie.TerminalCommitment {
	if len(data) == 0 {
		// empty slice -> no data (deleted)
		return nil
	}
	return hashData(data)
}

func (m *trieSetup) UpdateCommitment(prev *trie.VectorCommitment, delta trie.VectorCommitment) {
	*prev = delta
}

func (m *trieSetup) UpdateNodeCommitment(n *trie.Node) trie.VectorCommitment {
	if n == nil {
		// no node, no commitment
		return nil
	}
	n.Terminal = n.NewTerminal
	for i, child := range n.ModifiedChildren {
		c := m.UpdateNodeCommitment(child)
		if c != nil {
			n.Children[i] = c
		} else {
			// deletion
			delete(n.Children, i)
		}
	}
	n.ModifiedChildren = make(map[byte]*trie.Node)
	ret := m.CommitToNode(n)
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

func (s *hashCommitment) Equal(another trie.Commitment) bool {
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
