package trie

import (
	"bytes"
	"github.com/iotaledger/wasp/packages/util"
	"io"
)

type Commitment interface {
	Read(r io.Reader) error
	Write(w io.Writer)
	String() string
	Equal(Commitment) bool
}

type VectorCommitment interface {
	Commitment
}

type TerminalCommitment interface {
	Commitment
}

// Node is a node of the 25š+-ary verkle Trie
type Node struct {
	PathFragment []byte // can't be longer than 256 bytes
	Children     map[byte]VectorCommitment
	Terminal     TerminalCommitment
	// non-persistent
	NewTerminal      TerminalCommitment
	ModifiedChildren map[byte]*Node
}

const (
	hasTerminalValueFlag = 0x01
	hasChildrenFlag      = 0x02
)

func NewNode(pathFragment []byte) *Node {
	return &Node{
		PathFragment:     pathFragment,
		Children:         make(map[uint8]VectorCommitment),
		Terminal:         nil,
		ModifiedChildren: make(map[uint8]*Node),
	}
}

func NodeFromBytes(setup CommitmentLogic, data []byte) (*Node, error) {
	ret := NewNode(nil)
	if err := ret.Read(bytes.NewReader(data), setup); err != nil {
		return nil, err
	}
	ret.NewTerminal = ret.Terminal
	return ret, nil
}

func (n *Node) IsEmpty() bool {
	return len(n.Children) == 0 && len(n.ModifiedChildren) == 0 && n.Terminal == nil && n.NewTerminal == nil
}

func (n *Node) Write(w io.Writer) {
	_ = util.WriteBytes16(w, n.PathFragment)

	var smallFlags byte
	if n.Terminal != nil {
		smallFlags = hasTerminalValueFlag
	}
	// compress children flags 32 bytes (if any)
	var flags [32]byte
	for i := range n.Children {
		flags[i/8] |= 0x1 << (i % 8)
		smallFlags |= hasChildrenFlag
	}
	_ = util.WriteByte(w, smallFlags)
	// write terminal commitment if any
	if smallFlags&hasTerminalValueFlag != 0 {
		n.Terminal.Write(w)
	}
	// write child commitments if any
	if smallFlags&hasChildrenFlag != 0 {
		_, _ = w.Write(flags[:])
		for i := 0; i < 256; i++ {
			child, ok := n.Children[uint8(i)]
			if !ok {
				continue
			}
			child.Write(w)
		}
	}
}

func (n *Node) Read(r io.Reader, setup CommitmentLogic) error {
	var err error
	if n.PathFragment, err = util.ReadBytes16(r); err != nil {
		return err
	}
	var smallFlags byte
	if smallFlags, err = util.ReadByte(r); err != nil {
		return err
	}
	if smallFlags&hasTerminalValueFlag != 0 {
		n.Terminal = setup.NewTerminalCommitment()
		if err := n.Terminal.Read(r); err != nil {
			return err
		}
	} else {
		n.Terminal = nil
	}
	if smallFlags&hasChildrenFlag != 0 {
		var flags [32]byte
		if _, err := r.Read(flags[:]); err != nil {
			return err
		}
		for i := 0; i < 256; i++ {
			ib := uint8(i)
			if flags[i/8]&(0x1<<(i%8)) != 0 {
				n.Children[ib] = setup.NewVectorCommitment()
				if err := n.Children[ib].Read(r); err != nil {
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
