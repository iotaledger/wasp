package trie

import (
	"bytes"
	"encoding/hex"
	"fmt"
)

type terminalOption int

const (
	terminal = terminalOption(iota)
	noChild
	split
)

func (t *Trie) Update1(key []byte, value []byte) {
	c := t.model.CommitToData(value)
	rootNode, ok := t.GetNode(nil)
	if !ok {
		t.newTerminalNode(nil, key, c)
		return
	}
	node, opt, pos, childIndex := t.findTerminalNode(key, 0, rootNode)
	switch opt {
	case terminal:
		node.NewTerminal = c
	case noChild:
		assert(node.Children[childIndex] == nil, "node.Children[childIndex]==nil")
		node.ModifiedChildren[childIndex] = t.newTerminalNode(key[:pos], key[pos:], c)
	case split:
		// TODO
	}
}

func (t *Trie) findTerminalNode(path []byte, pathPosition int, node *Node) (*Node, terminalOption, int, byte) {
	for {
		assert(pathPosition <= len(path), "pathPosition<=len(path)")
		if bytes.Equal(path[pathPosition:], node.PathFragment) {
			return node, terminal, 0, 0
		}
		prefix := commonPrefix(path[pathPosition:], node.PathFragment)

		childIndexPosition := pathPosition + len(prefix)
		pathPosition = childIndexPosition + 1

		assert(childIndexPosition < len(path), "childIndexPosition<len(path)")
		childIndex := path[childIndexPosition]
		if len(prefix) != len(node.PathFragment) {
			return node, split, pathPosition, childIndex
		}
		child := node.Children[childIndex]
		if child == nil {
			return node, noChild, pathPosition, childIndex
		}
		var ok bool
		node, ok = t.GetNode(path[:pathPosition])
		if !ok {
			panic(fmt.Sprintf("inconsistency: trie key not found: %d", hex.EncodeToString(path[:childIndexPosition+1])))
		}
	}
}
