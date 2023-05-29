package trie

import "bytes"

// Diff computes the difference between two given trie roots, returning the collections
// of nodes that are exclusive to each trie.
func Diff(store KVStore, root1, root2 Hash) (onlyOn1, onlyOn2 map[Hash]*NodeData) {
	type nodeData struct {
		*NodeData
		key []byte
	}

	iterateTrie := func(tr *TrieReader) (*nodeData, func(IterateNodesAction) *nodeData) {
		nodes := make(chan *nodeData, 1)
		actions := make(chan IterateNodesAction, 1)

		go func() {
			defer close(nodes)
			tr.IterateNodes(func(nodeKey []byte, node *NodeData, depth int) IterateNodesAction {
				nodes <- &nodeData{NodeData: node, key: nodeKey}
				action := <-actions
				return action
			})
		}()

		firstNode := <-nodes
		next := func(a IterateNodesAction) *nodeData {
			actions <- a
			node, ok := <-nodes
			if !ok {
				actions <- IterateStop
				return nil
			}
			return node
		}
		return firstNode, next
	}

	tr1, err := NewTrieReader(store, root1)
	mustNoErr(err)
	tr2, err := NewTrieReader(store, root2)
	mustNoErr(err)
	current1, next1 := iterateTrie(tr1)
	current2, next2 := iterateTrie(tr2)

	onlyOn1 = make(map[Hash]*NodeData)
	onlyOn2 = make(map[Hash]*NodeData)

	// This is similar to the 'merge' function in mergeSort.
	// We iterate both tries in order, advancing the iterator of the smallest
	// node between the two.

	for current1 != nil && current2 != nil {
		if current1.Commitment == current2.Commitment {
			current1 = next1(IterateSkipSubtree)
			current2 = next2(IterateSkipSubtree)
		} else if bytes.Compare(current1.key, current2.key) < 0 {
			// TODO: can a node be "moved" (i.e. changed key but same commitment)?
			onlyOn1[current1.Commitment] = current1.NodeData
			current1 = next1(IterateContinue)
		} else {
			onlyOn2[current2.Commitment] = current2.NodeData
			current2 = next2(IterateContinue)
		}
	}
	for current1 != nil {
		onlyOn1[current1.Commitment] = current1.NodeData
		current1 = next1(IterateContinue)
	}
	for current2 != nil {
		onlyOn2[current2.Commitment] = current2.NodeData
		current2 = next2(IterateContinue)
	}
	return onlyOn1, onlyOn2
}
