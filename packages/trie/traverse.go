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

// pathElement proof element is NodeData together with the index of
// the next child in the path (except the last one in the proof path)
// Sequence of pathElement is used to generate proof
type pathElement struct {
	NodeData   *nodeData
	ChildIndex byte
}

// nodePath returns path PathElement-s along the triePath (the key) with the ending code
// to determine is it a proof of inclusion or absence
// Each path element contains index of the subsequent child, except the last one is set to 0
func (tr *TrieReader) nodePath(triePath []byte) ([]*pathElement, pathEndingCode) {
	ret := make([]*pathElement, 0)
	var endingCode pathEndingCode
	tr.traversePath(triePath, func(n *nodeData, trieKey []byte, ending pathEndingCode) {
		elem := &pathElement{
			NodeData: n,
		}
		nextChildIdx := len(trieKey) + len(n.pathExtension)
		if nextChildIdx < len(triePath) {
			elem.ChildIndex = triePath[nextChildIdx]
		}
		endingCode = ending
		ret = append(ret, elem)
	})
	assert(len(ret) > 0, "len(ret)>0")
	ret[len(ret)-1].ChildIndex = 0
	return ret, endingCode
}

func (tr *TrieReader) traversePath(target []byte, fun func(*nodeData, []byte, pathEndingCode)) {
	n, found := tr.nodeStore.FetchNodeData(tr.root)
	if !found {
		return
	}
	var path []byte
	for {
		pathPlusExtension := concat(path, n.pathExtension)
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

func (tr *TrieUpdatable) traverseMutatedPath(triePath []byte, fun func(n *bufferedNode, ending pathEndingCode)) {
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
			assert(len(keyPlusPathExtension) < len(triePath), "len(keyPlusPathExtension) < len(triePath)")
			prefix, _, _ := commonPrefix(keyPlusPathExtension, triePath)
			if !bytes.Equal(prefix, keyPlusPathExtension) {
				fun(n, endingSplit)
				return
			}
			childIndex := triePath[len(keyPlusPathExtension)]
			child := n.getChild(childIndex, tr.nodeStore)
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
