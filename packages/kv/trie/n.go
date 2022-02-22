package trie

import (
	"bytes"
	"encoding/hex"
	"fmt"
)

type endingCode byte

const (
	endingTerminal = iota
	endingSplit
	endingExtend
)

func (t *Trie) Update(key []byte, value []byte) {
	c := t.model.CommitToData(value)
	proof, lastKey, lastCommonPrefix, ending := t.path1(key, 0)
	if len(proof) == 0 {
		if c != nil {
			t.newTerminalNode(nil, key, c)
		}
		return

	}
	last := proof[len(proof)-1].Node
	switch ending {
	case endingTerminal:
		last.ModifiedTerminal = c

	case endingExtend:
		if c == nil {
			break
		}
		childPosition := len(lastKey) + len(lastCommonPrefix)
		assert(childPosition < len(key), "childPosition < len(key)")
		assert(last.Children[key[childPosition]] == nil, "last.Node.Children[key[childPosition]]")
		last.ModifiedChildren[key[childPosition]] = t.newTerminalNode(key[:childPosition+1], key[childPosition+1:], c)

	case endingSplit:
		if c == nil {
			break
		}
		childPosition := len(lastKey) + len(lastCommonPrefix)
		assert(childPosition <= len(key), "childPosition < len(key)")
		keyContinue := make([]byte, childPosition+1)
		copy(keyContinue, key)
		splitChildIndex := len(lastCommonPrefix)
		assert(splitChildIndex < len(last.PathFragment), "splitChildIndex<len(last.Node.PathFragment)")
		childContinue := last.PathFragment[splitChildIndex]
		keyContinue[len(keyContinue)-1] = childContinue

		// create new node on keyContinue, move everything from old to the new node and adjust the path fragment
		insertNode := t.newNodeCopy(keyContinue, last.PathFragment[splitChildIndex+1:], last)
		// clear the old one and adjust path fragment. Continue with 1 child, the new node
		last.Children = make(map[uint8]VCommitment)
		last.ModifiedChildren = make(map[uint8]*Node)
		last.PathFragment = lastCommonPrefix
		last.ModifiedChildren[childContinue] = insertNode
		last.Terminal = nil
		last.ModifiedTerminal = nil
		// insert terminal
		if childPosition == len(key) {
			// no need for the new node
			last.ModifiedTerminal = c
		} else {
			// create a new node
			keyFork := key[:len(keyContinue)]
			childForkIndex := keyFork[len(keyFork)-1]
			assert(int(childForkIndex) != splitChildIndex, "childForkIndex != splitChildIndex")
			last.ModifiedChildren[childForkIndex] = t.newTerminalNode(keyFork, key[len(keyFork):], c)
		}

	default:
		panic("inconsistency: unknown ending code")
	}
	for i := len(proof) - 2; i >= 0; i-- {
		k := proof[i+1].Key
		childIndex := k[len(k)-1]
		proof[i].Node.ModifiedChildren[childIndex] = proof[i+1].Node
	}
}

// returns key of the last node and common prefix with the fragment
func (t *Trie) path1(path []byte, pathPosition int) ([]ProofGenericElement, []byte, []byte, endingCode) {
	node, ok := t.GetNode(nil)
	if !ok {
		return nil, nil, nil, 0
	}

	proof := []ProofGenericElement{{Key: nil, Node: node}}
	key := path[:pathPosition]
	tail := path[pathPosition:]

	for {
		assert(pathPosition <= len(path), "pathPosition<=len(path)")
		if bytes.Equal(tail, node.PathFragment) {
			return proof, nil, nil, endingTerminal
		}
		assert(len(key) < len(path), "pathPosition<len(path)")
		prefix := commonPrefix(tail, node.PathFragment)

		if len(prefix) != len(node.PathFragment) {
			return proof, key, prefix, endingSplit
		}
		childIndexPosition := len(key) + len(prefix)
		assert(childIndexPosition < len(path), "childIndexPosition<len(path)")
		if node.Children[path[childIndexPosition]] == nil {
			return proof, key, prefix, endingExtend
		}

		pathPosition = childIndexPosition + 1
		key = path[:pathPosition]
		tail = path[pathPosition:]
		node, ok = t.GetNode(path[:pathPosition])
		if !ok {
			panic(fmt.Sprintf("inconsistency: trie key not found: %d", hex.EncodeToString(path[:childIndexPosition+1])))
		}
		proof = append(proof, ProofGenericElement{Key: key, Node: node})
	}
}
