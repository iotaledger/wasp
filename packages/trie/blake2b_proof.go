package trie

// MerkleProof is a proof of inclusion or absence
type MerkleProof struct {
	Key  []byte
	Path []*MerkleProofElement
}

type MerkleProofElement struct {
	PathExtension []byte
	Children      [NumChildren]*Hash
	Terminal      []byte
	ChildIndex    int
}

func (tr *TrieReader) MerkleProof(key []byte) *MerkleProof {
	unpackedKey := unpackBytes(key)
	nodePath, ending := tr.nodePath(unpackedKey)
	ret := &MerkleProof{
		Key:  unpackedKey,
		Path: make([]*MerkleProofElement, len(nodePath)),
	}
	for i, e := range nodePath {
		elem := &MerkleProofElement{
			PathExtension: e.NodeData.pathExtension,
			Terminal:      nil,
			ChildIndex:    int(e.ChildIndex),
		}
		if e.NodeData.terminal != nil {
			elem.Terminal, _ = compressToHashSize(e.NodeData.terminal.Bytes())
		}
		isLast := i == len(nodePath)-1
		for childIndex, childCommitment := range e.NodeData.children {
			if childCommitment == nil {
				continue
			}
			if !isLast && childIndex == int(e.ChildIndex) {
				// commitment to the next child is not included, it must be calculated by the verifier
				continue
			}
			hash := childCommitment.Hash()
			elem.Children[childIndex] = &hash
		}
		ret.Path[i] = elem
	}
	assert(len(ret.Path) > 0, "len(ret.Path)")
	last := ret.Path[len(ret.Path)-1]
	switch ending {
	case endingTerminal:
		last.ChildIndex = terminalCommitmentIndex
	case endingExtend, endingSplit:
		last.ChildIndex = pathExtensionCommitmentIndex
	default:
		panic("wrong ending code")
	}
	return ret
}
