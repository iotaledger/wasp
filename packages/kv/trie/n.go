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
	nodes := []*Node{rootNode}
	opt, pos, childIndex := t.path1(key, 0, nodes)
	node := nodes[len(nodes)-1]
	switch opt {
	case terminal:
		node.NewTerminal = c
	case noChild:
		node.ModifiedChildren[childIndex] = t.newTerminalNode(key[:pos], key[pos:], c)
	case split:
		// TODO
	}
}

func (t *Trie) path1(path []byte, pathPosition int, ret []*Node) (terminalOption, int, byte) {
	assert(len(ret) > 0, "len(ret)>0")
	assert(pathPosition <= len(path), "pathPosition<=len(path)")
	node := ret[len(ret)-1]
	if bytes.Equal(path[pathPosition:], node.PathFragment) {
		return terminal, 0, 0
	}
	prefix := commonPrefix(path[pathPosition:], node.PathFragment)
	nextPosition := pathPosition + len(prefix)
	assert(nextPosition < len(path), "nextPosition<len(path)")
	childIndex := path[nextPosition]
	if len(prefix) != len(node.PathFragment) {
		return split, nextPosition + 1, childIndex
	}
	child := node.Children[childIndex]
	if child == nil {
		return noChild, nextPosition + 1, childIndex
	}
	nextNode, ok := t.GetNode(path[:nextPosition+1])
	if !ok {
		panic(fmt.Sprintf("inconsistency: trie key not found: %d", hex.EncodeToString(path[:nextPosition+1])))
	}
	ret = append(ret, nextNode)
	return t.path1(path, nextPosition+1, ret)
}
