package trie

import (
	"bytes"
	"github.com/iotaledger/wasp/packages/util"
	"io"
)

type commitment interface {
	Read(r io.Reader) error
	Write(w io.Writer)
	String() string
	Equal(commitment) bool
}

type VectorCommitment interface {
	commitment
}

type TerminalCommitment interface {
	commitment
}

// Node is a node of the 25š+-ary verkle trie
type Node struct {
	pathFragment       []byte // can't be longer than 256 bytes
	children           map[byte]VectorCommitment
	terminalCommitment TerminalCommitment
	// non-persistent
	newTerminal      TerminalCommitment
	modifiedChildren map[byte]*Node
}

const (
	hasTerminalValueFlag = 0x01
	hasChildrenFlag      = 0x02
)

func NewNode() *Node {
	return &Node{
		pathFragment:       nil,
		children:           make(map[uint8]VectorCommitment),
		terminalCommitment: nil,
		modifiedChildren:   make(map[uint8]*Node),
	}
}

func (f *TrieSetup) NodeFromBytes(data []byte) (*Node, error) {
	ret := NewNode()
	if err := ret.Read(bytes.NewReader(data), f); err != nil {
		return nil, err
	}
	ret.newTerminal = ret.terminalCommitment
	return ret, nil
}

func (n *Node) IsEmpty() bool {
	return len(n.children) == 0 && len(n.modifiedChildren) == 0 && n.terminalCommitment == nil && n.newTerminal == nil
}

func (n *Node) Write(w io.Writer) {
	_ = util.WriteBytes16(w, n.pathFragment)

	var smallFlags byte
	if n.terminalCommitment != nil {
		smallFlags = hasTerminalValueFlag
	}
	// compress children flags 32 bytes (if any)
	var flags [32]byte
	for i := range n.children {
		flags[i/8] |= 0x1 << (i % 8)
		smallFlags |= hasChildrenFlag
	}
	_ = util.WriteByte(w, smallFlags)
	// write terminal commitment if any
	if smallFlags&hasTerminalValueFlag != 0 {
		n.terminalCommitment.Write(w)
	}
	// write child commitments if any
	if smallFlags&hasChildrenFlag != 0 {
		_, _ = w.Write(flags[:])
		for i := 0; i < 256; i++ {
			child, ok := n.children[uint8(i)]
			if !ok {
				continue
			}
			child.Write(w)
		}
	}
}

func (n *Node) Read(r io.Reader, factory *TrieSetup) error {
	var err error
	if n.pathFragment, err = util.ReadBytes16(r); err != nil {
		return err
	}
	var smallFlags byte
	if smallFlags, err = util.ReadByte(r); err != nil {
		return err
	}
	if smallFlags&hasTerminalValueFlag != 0 {
		n.terminalCommitment = factory.NewTerminalCommitment()
		if err := n.terminalCommitment.Read(r); err != nil {
			return err
		}
	} else {
		n.terminalCommitment = nil
	}
	if smallFlags&hasChildrenFlag != 0 {
		var flags [32]byte
		if _, err := r.Read(flags[:]); err != nil {
			return err
		}
		for i := 0; i < 256; i++ {
			ib := uint8(i)
			if flags[i/8]&(0x1<<(i%8)) != 0 {
				n.children[ib] = factory.NewVectorCommitment()
				if err := n.children[ib].Read(r); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func Bytes(o interface{ Write(w io.Writer) }) []byte {
	var buf bytes.Buffer
	o.Write(&buf)
	return buf.Bytes()
}

type byteCounter int

func (b *byteCounter) Write(p []byte) (n int, err error) {
	*b = byteCounter(int(*b) + len(p))
	return 0, nil
}

func Size(o interface{ Write(w io.Writer) }) int {
	var ret byteCounter
	o.Write(&ret)
	return int(ret)
}
