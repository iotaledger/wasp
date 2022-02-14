package merkle

import (
	"bytes"
	"github.com/iotaledger/wasp/packages/kv/trie"
	"golang.org/x/xerrors"
)

type MerkleProof struct {
	Key  []byte
	Path []*MerkleProofElement
}

type MerkleProofElement struct {
	PathFragment []byte
	Children     map[byte]*[32]byte
	Terminal     *[32]byte
	ChildIndex   int
}

// ProvePath converts generic proof path to the Merkle proof path
func (m *merkleTrieSetup) ProvePath(path *trie.ProofPath) *MerkleProof {
	ret := &MerkleProof{
		Key:  path.Key,
		Path: make([]*MerkleProofElement, len(path.Path)),
	}
	for i, eg := range path.Path {
		em := &MerkleProofElement{
			PathFragment: eg.Node.PathFragment,
			Children:     make(map[byte]*[32]byte),
			Terminal:     nil,
			ChildIndex:   eg.ChildIndex,
		}
		if eg.Node.Terminal != nil {
			em.Terminal = (*[32]byte)(eg.Node.Terminal.(*hashCommitment))
		}
		for k, v := range eg.Node.Children {
			if int(k) == eg.ChildIndex {
				// skipping the commitment which must come from the next child. 256 and 257 will be skipped too
				continue
			}
			em.Children[k] = (*[32]byte)(v.(*hashCommitment))
		}
		ret.Path[i] = em
	}
	return ret
}

// KeyTerminal returns key and terminal commitment the proof is about. If it returns (?, nil) it means it is proof of absence
// It does not verify the proof, so this function should be used only after Verify()
func (p *MerkleProof) KeyTerminal() ([]byte, *[32]byte) {
	if len(p.Path) == 0 {
		return nil, nil
	}
	lastElem := p.Path[len(p.Path)-1]
	if lastElem.ChildIndex == 257 {
		return p.Key, nil
	}
	return p.Key, lastElem.Terminal
}

func (p *MerkleProof) Verify(root *[32]byte) error {
	if len(p.Path) == 0 {
		if root != nil {
			return xerrors.New("proof is empty")
		}
		return nil
	}
	c, err := p.verify(0, 0)
	if err != nil {
		return err
	}
	if c != *root {
		return xerrors.New("commitment not equal to the root")
	}
	return nil
}

func (p *MerkleProof) verify(pathIdx, keyIdx int) ([32]byte, error) {
	assert(pathIdx < len(p.Path), "assertion: pathIdx < len(p.Path)")
	assert(keyIdx < len(p.Key), "assertion: keyIdx < len(p.Key)")

	elem := p.Path[pathIdx]
	tail := p.Key[keyIdx:]
	isPrefix := bytes.HasPrefix(tail, elem.PathFragment)
	last := pathIdx == len(p.Path)-1
	if !last && !isPrefix {
		return [32]byte{}, xerrors.Errorf("wrong proof: proof path does not follow the key. Path position: %d, key position %d", pathIdx, keyIdx)
	}
	if !last {
		assert(isPrefix, "assertion: isPrefix")
		if elem.ChildIndex > 255 {
			return [32]byte{}, xerrors.Errorf("wrong proof: wrong child index. Path position: %d, key position %d", pathIdx, keyIdx)
		}
		if elem.Children[byte(elem.ChildIndex)] != nil {
			return [32]byte{}, xerrors.Errorf("wrong proof: nil expected at child index. Path position: %d, key position %d", pathIdx, keyIdx)
		}
		nextKeyIdx := keyIdx + len(elem.PathFragment) + 1
		if nextKeyIdx > len(p.Key) {
			return [32]byte{}, xerrors.Errorf("wrong proof: proof path out of key bounds. Path position: %d, key position %d", pathIdx, keyIdx)
		}
		c, err := p.verify(pathIdx+1, nextKeyIdx)
		if err != nil {
			return [32]byte{}, err
		}
		return elem.hashIt(&c), nil
	}
	// it is the last in the path
	if elem.ChildIndex != 256 && elem.ChildIndex != 257 {
		return [32]byte{}, xerrors.Errorf("wrong proof: child index expected to be 256 or 257. Path position: %d, key position %d", pathIdx, keyIdx)
	}
	return elem.hashIt(nil), nil
}

func (e *MerkleProofElement) hashIt(missingCommitment *[32]byte) [32]byte {
	var hashes [258]*hashCommitment
	for idx, c := range e.Children {
		hashes[idx] = (*hashCommitment)(c)
	}
	hashes[256] = (*hashCommitment)(e.Terminal)
	hashes[257] = hashData(e.PathFragment)
	if e.ChildIndex < 256 {
		hashes[e.ChildIndex] = (*hashCommitment)(missingCommitment)
	}
	return *hashHashes(&hashes)
}

func assert(cond bool, err interface{}) {
	if !cond {
		panic(err)
	}
}
