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
	rootNode, ok := t.GetNode(nil)
	if !ok {
		if c != nil {
			t.newTerminalNode(nil, key, c)
		}
		return
	}
	proof := []ProofGenericElement{{Node: rootNode}}
	lastKey, lastCommonPrefix, ending := t.path1(key, 0, proof)

	last := proof[len(proof)-1]
	switch ending {
	case endingTerminal:
		last.Node.ModifiedTerminal = c

	case endingExtend:
		if c == nil {
			break
		}
		childPosition := len(lastKey) + len(lastCommonPrefix)
		assert(childPosition < len(key), "childPosition < len(key)")
		assert(last.Node.Children[key[childPosition]] == nil, "last.Node.Children[key[childPosition]]")
		last.Node.ModifiedChildren[key[childPosition]] = t.newTerminalNode(key[:childPosition+1], key[childPosition+1:], c)

	case endingSplit:
		if c == nil {
			break
		}
		childPosition := len(lastKey) + len(lastCommonPrefix)
		assert(childPosition <= len(key), "childPosition < len(key)")
		keyContinue := make([]byte, childPosition+1)
		copy(keyContinue, key)
		splitChildIndex := len(lastCommonPrefix)
		assert(splitChildIndex < len(last.Node.PathFragment), "splitChildIndex<len(last.Node.PathFragment)")
		childContinue := last.Node.PathFragment[splitChildIndex]
		keyContinue[len(keyContinue)-1] = childContinue

		// create new node on keyContinue, move everything from old to the new node and adjust the path fragment
		insertNode := t.newNodeCopy(keyContinue, last.Node.PathFragment[splitChildIndex+1:], last.Node)
		// clear the old one and adjust path fragment. Continue with 1 child, the new node
		last.Node.Children = make(map[uint8]VCommitment)
		last.Node.ModifiedChildren = make(map[uint8]*Node)
		last.Node.PathFragment = lastCommonPrefix
		last.Node.ModifiedChildren[childContinue] = insertNode
		last.Node.Terminal = nil
		last.Node.ModifiedTerminal = nil
		// insert terminal
		if childPosition == len(key) {
			// no need for the new node
			last.Node.ModifiedTerminal = c
		} else {
			// create a new node
			keyFork := key[:len(keyContinue)]
			childForkIndex := keyFork[len(keyFork)-1]
			assert(int(childForkIndex) != splitChildIndex, "childForkIndex != splitChildIndex")
			last.Node.ModifiedChildren[childForkIndex] = t.newTerminalNode(keyFork, key[len(keyFork):], c)
		}

	default:
		panic("inconsistency: unknown ending code")
	}
	for i := len(proof) - 2; i >= 0; i-- {
		proof[i].Node.ModifiedChildren[proof[i].ChildIndex] = proof[i+1].Node
	}
}

// returns key of the last node and common prefix with the fragment
func (t *Trie) path1(path []byte, pathPosition int, proof []ProofGenericElement) ([]byte, []byte, endingCode) {
	assert(len(proof) == 1, "len(proof)==1")
	elem := &proof[0]
	for {
		assert(pathPosition <= len(path), "pathPosition<=len(path)")
		key := path[:pathPosition]
		tail := path[pathPosition:]
		if bytes.Equal(tail, elem.Node.PathFragment) {
			return nil, nil, endingTerminal
		}
		assert(len(key) < len(path), "pathPosition<len(path)")
		prefix := commonPrefix(tail, elem.Node.PathFragment)

		if len(prefix) != len(elem.Node.PathFragment) {
			return key, prefix, endingSplit
		}
		childIndexPosition := len(key) + len(prefix)
		assert(childIndexPosition < len(path), "childIndexPosition<len(path)")
		if elem.Node.Children[path[childIndexPosition]] == nil {
			return key, prefix, endingExtend
		}
		pathPosition = childIndexPosition + 1
		node, ok := t.GetNode(path[:pathPosition])
		if !ok {
			panic(fmt.Sprintf("inconsistency: trie key not found: %d", hex.EncodeToString(path[:childIndexPosition+1])))
		}
		elem = &ProofGenericElement{Node: node}
		proof = append(proof, *elem)
	}
}
