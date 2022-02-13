package trie

type Proof interface {
}

func (t *trie) Prove(key []byte) Proof {
	return nil
}
