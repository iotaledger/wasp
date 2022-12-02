package trie

import (
	"bytes"
	"fmt"
	"io"
)

const (
	// NumChildren is the maximum amount of children for each trie node
	NumChildren = 16
)

func isValidChildIndex(i int) bool {
	return i >= 0 && i < NumChildren
}

// nodeData represents a node of the trie, which is stored in the trieStore
// with key = commitment.Bytes()
type nodeData struct {
	// if pathExtension != nil, this is an extension node (i.e. if there are
	// no branching nodes along the pathExtension).
	// See https://ethereum.org/en/developers/docs/data-structures-and-encoding/patricia-merkle-trie/#optimization
	pathExtension []byte

	// if terminal != nil, it contains the commitment to a value in the trie
	terminal TCommitment

	// children contains pointers to up to 16 other nodes, one for each
	// possible nibble
	children [NumChildren]VCommitment

	// commitment is the hash(node), which is persisted in the key
	commitment VCommitment
}

func newNodeData() *nodeData {
	n := &nodeData{}
	v := vectorCommitment(makeHashVector(n).Hash())
	n.commitment = &v
	return n
}

func nodeDataFromBytes(data []byte) (*nodeData, error) {
	ret := newNodeData()
	rdr := bytes.NewReader(data)
	if err := ret.Read(rdr); err != nil {
		return nil, err
	}
	if rdr.Len() != 0 {
		// not all data was consumed
		return nil, ErrNotAllBytesConsumed
	}
	return ret, nil
}

func (n *nodeData) ChildrenCount() int {
	count := 0
	for _, c := range n.children {
		if c != nil {
			count++
		}
	}
	return count
}

// Clone deep copy
func (n *nodeData) Clone() *nodeData {
	ret := &nodeData{
		pathExtension: concat(n.pathExtension),
	}
	if n.terminal != nil {
		ret.terminal = n.terminal.Clone()
	}
	if n.commitment != nil {
		ret.commitment = n.commitment.Clone()
	}
	for i, c := range n.children {
		if c != nil {
			ret.children[i] = c.Clone()
		}
	}
	return ret
}

func (n *nodeData) String() string {
	t := "<nil>"
	if n.terminal != nil {
		t = n.terminal.String()
	}
	childIdx := make([]byte, 0)
	for i := range n.children {
		if n.children[i] != nil {
			childIdx = append(childIdx, byte(i))
		}
	}
	return fmt.Sprintf("c: %s, pf: '%s', childrenIdx: %v, term: '%s'",
		n.commitment, string(n.pathExtension), childIdx, t)
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

// cflags 16 flags, one for each child
type cflags uint16

func readCflags(r io.Reader) (cflags, error) {
	var ret uint16
	err := readUint16(r, &ret)
	if err != nil {
		return 0, err
	}
	return cflags(ret), nil
}

func (fl *cflags) setFlag(i byte) {
	*fl |= 0x1 << i
}

func (fl cflags) hasFlag(i byte) bool {
	return fl&(0x1<<i) != 0
}

// Write serialized node data
func (n *nodeData) Write(w io.Writer) error {
	var smallFlags byte
	if n.terminal != nil {
		smallFlags |= isTerminalNodeFlag
	}

	childrenFlags := cflags(0)
	// compress children childrenFlags 32 bytes, if any
	for i := range n.children {
		if n.children[i] != nil {
			childrenFlags.setFlag(byte(i))
		}
	}

	if childrenFlags != 0 {
		smallFlags |= hasChildrenFlag
	}
	var pathExtensionEncoded []byte
	var err error
	if len(n.pathExtension) > 0 {
		smallFlags |= isExtensionNodeFlag
		if pathExtensionEncoded, err = encodeUnpackedBytes(n.pathExtension); err != nil {
			return err
		}
	}
	if err = writeByte(w, smallFlags); err != nil {
		return err
	}
	if smallFlags&isExtensionNodeFlag != 0 {
		if err = writeBytes16(w, pathExtensionEncoded); err != nil {
			return err
		}
	}
	if smallFlags&isTerminalNodeFlag != 0 {
		if err = n.terminal.Write(w); err != nil {
			return err
		}
	}
	// write child commitments if any
	if smallFlags&hasChildrenFlag != 0 {
		if err = writeUint16(w, uint16(childrenFlags)); err != nil {
			return err
		}
		for _, child := range n.children {
			if child != nil {
				if err = child.Write(w); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Read deserialize node data
func (n *nodeData) Read(r io.Reader) error {
	var err error
	var smallFlags byte
	if smallFlags, err = readByte(r); err != nil {
		return err
	}
	if smallFlags&isExtensionNodeFlag != 0 {
		encoded, err := readBytes16(r)
		if err != nil {
			return err
		}
		if n.pathExtension, err = decodeToUnpackedBytes(encoded); err != nil {
			return err
		}
	} else {
		n.pathExtension = nil
	}
	n.terminal = nil
	if smallFlags&isTerminalNodeFlag != 0 {
		n.terminal = newTerminalCommitment()
		if err = n.terminal.Read(r); err != nil {
			return err
		}
	}
	if smallFlags&hasChildrenFlag != 0 {
		var flags cflags
		if flags, err = readCflags(r); err != nil {
			return err
		}
		for i := 0; i < NumChildren; i++ {
			ib := uint8(i)
			if flags.hasFlag(ib) {
				n.children[ib] = newVectorCommitment()
				if err = n.children[ib].Read(r); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (n *nodeData) iterateChildren(f func(byte, VCommitment) bool) bool {
	for i, v := range n.children {
		if v != nil {
			if !f(byte(i), v) {
				return false
			}
		}
	}
	return true
}
