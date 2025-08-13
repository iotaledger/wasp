package trie

import (
	"bytes"
)

// pathEndingCode is a tag how trie path ends wrt the trieKey
type pathEndingCode byte

const (
	endingNone = pathEndingCode(iota)
	endingTerminal
	endingSplit
	endingExtend
)

func (e pathEndingCode) String() string {
	switch e {
	case endingNone:
		return "EndingNone"
	case endingTerminal:
		return "EndingTerminal"
	case endingSplit:
		return "EndingSplit"
	case endingExtend:
		return "EndingExtend"
	default:
		panic("invalid ending code")
	}
}

func (tr *Reader) traversePath(target []byte, fun func(*NodeData, []byte, pathEndingCode)) {
	n, found := tr.nodeStore.FetchNodeData(tr.root)
	if !found {
		return
	}
	var path []byte
	for {
		pathPlusExtension := concat(path, n.PathExtension)
		switch {
		case len(pathPlusExtension) > len(target):
			fun(n, path, endingSplit)
			return
		case len(pathPlusExtension) == len(target):
			if bytes.Equal(pathPlusExtension, target) {
				fun(n, path, endingTerminal)
			} else {
				fun(n, path, endingSplit)
			}
			return
		default:
			prefix, _, _ := commonPrefix(pathPlusExtension, target)
			if !bytes.Equal(prefix, pathPlusExtension) {
				fun(n, path, endingSplit)
				return
			}
			childIndex := target[len(pathPlusExtension)]
			child, childTrieKey := tr.nodeStore.FetchChild(n, childIndex, path)
			if child == nil {
				fun(n, childTrieKey, endingExtend)
				return
			}
			fun(n, path, endingNone)
			path = childTrieKey
			n = child
		}
	}
}

func (tr *Draft) traverseMutatedPath(triePath []byte, fun func(n *draftNode, ending pathEndingCode)) {
	n := tr.mutatedRoot
	for {
		keyPlusPathExtension := concat(n.triePath, n.pathExtension)
		switch {
		case len(triePath) < len(keyPlusPathExtension):
			fun(n, endingSplit)
			return
		case len(triePath) == len(keyPlusPathExtension):
			if bytes.Equal(keyPlusPathExtension, triePath) {
				fun(n, endingTerminal)
			} else {
				fun(n, endingSplit)
			}
			return
		default:
			assertf(len(keyPlusPathExtension) < len(triePath), "len(keyPlusPathExtension) < len(triePath)")
			prefix, _, _ := commonPrefix(keyPlusPathExtension, triePath)
			if !bytes.Equal(prefix, keyPlusPathExtension) {
				fun(n, endingSplit)
				return
			}
			childIndex := triePath[len(keyPlusPathExtension)]
			child := n.getChild(childIndex, tr.base.nodeStore)
			if child == nil {
				fun(n, endingExtend)
				return
			}
			fun(n, endingNone)
			n = child
		}
	}
}

func commonPrefix(b1, b2 []byte) (prefix []byte, tail1 []byte, tail2 []byte) {
	i := 0
	for ; i < len(b1) && i < len(b2); i++ {
		if b1[i] != b2[i] {
			break
		}
	}
	prefix = b1[:i]
	tail1 = b1[i:]
	tail2 = b2[i:]
	return
}
