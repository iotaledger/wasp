package trie

import (
	"bytes"
	"encoding/hex"
	"fmt"
)

const (
	terminalEnding = 256
	splitEnding    = 257
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
	lastKey, lastCommonPrefix := t.path1(key, 0, proof)

	last := proof[len(proof)-1]
	pathPosition := len(lastKey) + len(lastCommonPrefix)
	switch {
	case last.ChildIndex < 256:
		assert(last.Node.Children[byte(last.ChildIndex)] == nil, "last.Node.Children[byte(last.ChildIndex)] == nil")
		if c != nil {
			last.Node.ModifiedChildren[byte(last.ChildIndex)] = t.newTerminalNode(key[:pathPosition], key[pathPosition:], c)
		}

	case last.ChildIndex == terminalEnding:
		last.Node.ModifiedTerminal = c

	case last.ChildIndex == splitEnding:
		if c == nil {
			break
		}
		keyToContinue := make([]byte, pathPosition+1)
		copy(keyToContinue, key)
		splitChildIndex := len(lastCommonPrefix)
		keyToContinue[pathPosition] = last.Node.PathFragment[splitChildIndex]

		// create new node on keyContinue, move everything from old to the new node and adjust the path fragment
		insertNode := t.newNodeCopy(keyToContinue, last.Node.PathFragment[len(lastCommonPrefix)+1:], last.Node)
		// clear the old one and adjust path fragment. Continue with 1 child, the new node
		last.Node.Children = make(map[uint8]VCommitment)
		last.Node.ModifiedChildren = make(map[uint8]*Node)
		last.Node.PathFragment = lastCommonPrefix
		last.Node.ModifiedChildren[byte(splitChildIndex)] = insertNode
		last.Node.Terminal = nil
		last.Node.ModifiedTerminal = nil
		// insert terminal
		if pathPosition+len(lastCommonPrefix) == len(key) {
			// no need for the new node
			last.Node.ModifiedTerminal = c
		} else {
			// create the new node
			keyFork := key[:pathPosition+1]
			childForkIndex := keyFork[len(keyFork)-1]
			assert(int(childForkIndex) != splitChildIndex, "childForkIndex != splitChildIndex")
			assert(len(keyToContinue) == len(keyFork), "len(keyContinue)==len(keyFork)")
			last.Node.ModifiedChildren[childForkIndex] = t.newTerminalNode(keyFork, key[len(keyFork):], c)
		}

	default:
		panic("expected <= 257")
	}
	for i := len(proof) - 2; i >= 0; i-- {
		idx := proof[i].ChildIndex
		assert(0 <= idx && idx < 256, "child index must be between 0 and 255")
		proof[i].Node.ModifiedChildren[byte(idx)] = proof[i+1].Node
	}
}

// returns key of the last node and common prefix with the fragment
func (t *Trie) path1(path []byte, pathPosition int, proof []ProofGenericElement) ([]byte, []byte) {
	assert(len(proof) == 1, "len(proof)==1")
	elem := proof[0]
	for {
		assert(pathPosition <= len(path), "pathPosition<=len(path)")
		if bytes.Equal(path[pathPosition:], elem.Node.PathFragment) {
			elem.ChildIndex = terminalEnding
			return path[:pathPosition], path[pathPosition:]
		}
		prefix := commonPrefix(path[pathPosition:], elem.Node.PathFragment)
		childIndexPosition := pathPosition + len(prefix)

		assert(childIndexPosition < len(path), "childIndexPosition<len(path)")
		childIndex := path[childIndexPosition]
		if len(prefix) != len(elem.Node.PathFragment) {
			elem.ChildIndex = int(childIndex)
			return path[:pathPosition], prefix
		}
		if elem.Node.Children[childIndex] == nil {
			elem.ChildIndex = splitEnding
			return path[:pathPosition], prefix
		}
		pathPosition = childIndexPosition + 1
		node, ok := t.GetNode(path[:pathPosition])
		if !ok {
			panic(fmt.Sprintf("inconsistency: trie key not found: %d", hex.EncodeToString(path[:childIndexPosition+1])))
		}
		elem = ProofGenericElement{Node: node}
		proof = append(proof, elem)
	}
}
