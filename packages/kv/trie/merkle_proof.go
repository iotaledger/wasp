package trie

type merkleProofElement struct {
	children  map[byte]hashCommitment
	terminal  hashCommitment
	pathIndex byte
}

func (m merkleProofElement) Proves(commitment VectorCommitment) {
	//TODO implement me
	panic("implement me")
}
