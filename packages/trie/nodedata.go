package trie

import (
	"fmt"
	"io"

	"fortio.org/safecast"

	"github.com/iotaledger/wasp/packages/util/rwutil"
)

const (
	// NumChildren is the maximum amount of children for each trie node
	NumChildren = 16
)

func isValidChildIndex(i int) bool {
	return i >= 0 && i < NumChildren
}

// NodeData represents a node of the trie, which is stored in the trieStore
// with key = commitment.Bytes()
type NodeData struct {
	// if PathExtension != nil, this is an extension node (i.e. if there are
	// no branching nodes along the PathExtension).
	// See https://ethereum.org/en/developers/docs/data-structures-and-encoding/patricia-merkle-trie/#optimization
	PathExtension []byte

	// if Terminal != nil, it contains the commitment to a value in the trie
	Terminal *Tcommitment

	// Children contains pointers to up to 16 other nodes, one for each
	// possible nibble
	Children [NumChildren]*Hash

	// Commitment is hash(pathExtension|terminal|children), which is persisted in the key
	Commitment Hash
}

func newNodeData() *NodeData {
	n := &NodeData{}
	n.updateCommitment()
	return n
}

func nodeDataFromBytes(data []byte) (*NodeData, error) {
	return rwutil.ReadFromBytes(data, newNodeData())
}

func (n *NodeData) Bytes() []byte {
	return rwutil.WriteToBytes(n)
}

func (n *NodeData) ChildrenCount() int {
	count := 0
	for _, c := range n.Children {
		if c != nil {
			count++
		}
	}
	return count
}

// Clone deep copy
func (n *NodeData) Clone() *NodeData {
	ret := &NodeData{
		PathExtension: concat(n.PathExtension),
	}
	if n.Terminal != nil {
		ret.Terminal = n.Terminal.Clone()
	}
	ret.Commitment = n.Commitment.Clone()
	n.iterateChildren(func(i byte, h Hash) bool {
		ret.Children[i] = &h
		return true
	})
	return ret
}

func (n *NodeData) String() string {
	t := "<nil>"
	if n.Terminal != nil {
		t = n.Terminal.String()
	}
	return fmt.Sprintf("c:%s ext:%v term:%s",
		n.Commitment, n.PathExtension, t)
}

func (n *NodeData) iterateChildren(f func(byte, Hash) bool) bool {
	for i, v := range n.Children {
		if v != nil {
			if !f(byte(i), *v) {
				return false
			}
		}
	}
	return true
}

// update computes update to the node data and its commitment.
func (n *NodeData) update(childUpdates map[byte]*Hash, newTerminalUpdate *Tcommitment, pathExtension []byte) {
	for i, upd := range childUpdates {
		n.Children[i] = upd
	}
	n.Terminal = newTerminalUpdate // for hash commitment just replace
	n.PathExtension = pathExtension
	n.updateCommitment()
}

func (n *NodeData) updateCommitment() {
	hashes := &hashVector{}
	n.iterateChildren(func(i byte, h Hash) bool {
		hashes[i] = h[:]
		return true
	})
	if n.Terminal != nil {
		// squeeze terminal it into the hash size, if longer than hash size
		hashes[terminalIndex] = compressToHashSize(n.Terminal.Bytes())
	}
	pathExtensionCommitmentBytes := compressToHashSize(n.PathExtension)
	hashes[pathExtensionIndex] = pathExtensionCommitmentBytes
	n.Commitment = hashes.Hash()
}

// Read/Write implements optimized serialization of the trie node
// The serialization of the node takes advantage of the fact that most of the
// nodes has just few children.
// the 'smallFlags' (1 byte) contains information:
// - 'hasChildrenFlag' does node contain at least one child
// - 'isTerminalNodeFlag' means that the node contains a terminal commitment
// - 'isExtensionNodeFlag' means that the node has a non-empty path extension
// By the semantics of the trie, 'smallFlags' cannot be 0
// 'childrenFlags' (2 bytes array or 16 bits) are only present if node contains at least one child commitment
// In this case:
// if node has a child commitment at the position of i, 0 <= p <= 255, it has a bit in the byte array
// at the index i/8. The bit position in the byte is i % 8

const (
	isTerminalNodeFlag = 1 << iota
	hasChildrenFlag
	isExtensionNodeFlag
)

func (n *NodeData) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	smallFlags := rr.ReadByte()
	n.PathExtension = nil
	if smallFlags&isExtensionNodeFlag != 0 {
		encoded := rr.ReadBytes()
		if rr.Err == nil {
			n.PathExtension, rr.Err = decodeToUnpackedBytes(encoded)
		}
	}
	n.Terminal = nil
	if smallFlags&isTerminalNodeFlag != 0 {
		n.Terminal = newTerminalCommitment()
		rr.Read(n.Terminal)
	}
	if smallFlags&hasChildrenFlag != 0 {
		flags := rr.ReadUint16()
		for i := 0; i < NumChildren; i++ {
			ib, err := safecast.Convert[uint8](i)
			if err != nil {
				panic(fmt.Sprintf("index %d is too large for uint8", i))
			}
			if (flags & (1 << i)) != 0 {
				n.Children[ib] = &Hash{}
				rr.Read(n.Children[ib])
			}
		}
	}
	return rr.Err
}

func (n *NodeData) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	var smallFlags byte
	if n.Terminal != nil {
		smallFlags |= isTerminalNodeFlag
	}

	// compress child indexes in 32 bits
	childrenFlags := uint16(0)
	n.iterateChildren(func(i byte, _ Hash) bool {
		childrenFlags |= 1 << i
		return true
	})
	if childrenFlags != 0 {
		smallFlags |= hasChildrenFlag
	}

	var pathExtensionEncoded []byte
	if len(n.PathExtension) > 0 {
		smallFlags |= isExtensionNodeFlag
		pathExtensionEncoded, ww.Err = encodeUnpackedBytes(n.PathExtension)
	}

	ww.WriteByte(smallFlags)
	if smallFlags&isExtensionNodeFlag != 0 {
		ww.WriteBytes(pathExtensionEncoded)
	}
	if smallFlags&isTerminalNodeFlag != 0 {
		ww.Write(n.Terminal)
	}
	if smallFlags&hasChildrenFlag != 0 {
		ww.WriteUint16(childrenFlags)
		n.iterateChildren(func(_ byte, h Hash) bool {
			ww.Write(h)
			return ww.Err == nil
		})
	}
	return ww.Err
}
