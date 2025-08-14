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

func (tr *TrieRFromRoot) MerkleProof(key []byte) *MerkleProof {
	unpackedKey := unpackBytes(key)
	nodePath, ending := tr.nodePath(unpackedKey)
	ret := &MerkleProof{
		Key:  unpackedKey,
		Path: make([]*MerkleProofElement, len(nodePath)),
	}
	for i, e := range nodePath {
		elem := &MerkleProofElement{
			PathExtension: e.NodeData.PathExtension,
			Terminal:      nil,
			ChildIndex:    int(e.ChildIndex),
		}
		if e.NodeData.Terminal != nil {
			elem.Terminal = compressToHashSize(e.NodeData.Terminal.Bytes())
		}
		isLast := i == len(nodePath)-1
		for childIndex, childCommitment := range e.NodeData.Children {
			if childCommitment == nil {
				continue
			}
			if !isLast && childIndex == int(e.ChildIndex) {
				// commitment to the next child is not included, it must be calculated by the verifier
				continue
			}
			elem.Children[childIndex] = childCommitment
		}
		ret.Path[i] = elem
	}
	assertf(len(ret.Path) > 0, "len(ret.Path)")
	last := ret.Path[len(ret.Path)-1]
	switch ending {
	case endingTerminal:
		last.ChildIndex = terminalIndex
	case endingExtend, endingSplit:
		last.ChildIndex = pathExtensionIndex
	default:
		panic("wrong ending code")
	}
	return ret
}

// pathElement proof element is NodeData together with the index of
// the next child in the path (except the last one in the proof path)
// Sequence of pathElement is used to generate proof
type pathElement struct {
	NodeData   *NodeData
	ChildIndex byte
}

// nodePath returns path PathElement-s along the triePath (the key) with the ending code
// to determine is it a proof of inclusion or absence
// Each path element contains index of the subsequent child, except the last one is set to 0
func (tr *TrieRFromRoot) nodePath(triePath []byte) ([]*pathElement, pathEndingCode) {
	ret := make([]*pathElement, 0)
	var endingCode pathEndingCode
	tr.R.traversePath(tr.Root, triePath, func(n *NodeData, trieKey []byte, ending pathEndingCode) {
		elem := &pathElement{
			NodeData: n,
		}
		nextChildIdx := len(trieKey) + len(n.PathExtension)
		if nextChildIdx < len(triePath) {
			elem.ChildIndex = triePath[nextChildIdx]
		}
		endingCode = ending
		ret = append(ret, elem)
	})
	assertf(len(ret) > 0, "len(ret)>0")
	ret[len(ret)-1].ChildIndex = 0
	return ret, endingCode
}
