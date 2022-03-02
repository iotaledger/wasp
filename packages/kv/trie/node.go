package trie

import (
	"bytes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/util"
	"io"
)

type CommitmentBase interface {
	Read(r io.Reader) error
	Write(w io.Writer) error
	String() string
	Bytes() []byte
}

type VCommitment interface {
	CommitmentBase
	Update(delta VCommitment)
	Clone() VCommitment
}

type TCommitment interface {
	CommitmentBase
	Clone() TCommitment
}

// Node is a node of the 256+-ary verkle Trie
type Node struct {
	PathFragment     []byte
	ChildCommitments map[byte]VCommitment
	Terminal         TCommitment
	// non-persistent, used for caching
	newTerminal      TCommitment
	modifiedChildren map[uint8]struct{} // to be updated child commitments
}

const (
	hasTerminalValueFlag = 0x01
	hasChildrenFlag      = 0x02
)

func NewNode(pathFragment []byte) *Node {
	return &Node{
		PathFragment:     pathFragment,
		ChildCommitments: make(map[uint8]VCommitment),
		Terminal:         nil,
		modifiedChildren: make(map[uint8]struct{}),
	}
}

func NodeFromBytes(setup CommitmentModel, data []byte) (*Node, error) {
	ret := NewNode(nil)
	if err := ret.Read(bytes.NewReader(data), setup); err != nil {
		return nil, err
	}
	ret.newTerminal = ret.Terminal
	return ret, nil
}

func (n *Node) Clone() *Node {
	if n == nil {
		return nil
	}
	ret := &Node{
		PathFragment:     make([]byte, len(n.PathFragment)),
		ChildCommitments: make(map[byte]VCommitment),
		Terminal:         n.Terminal.Clone(),
		newTerminal:      n.newTerminal.Clone(),
		modifiedChildren: make(map[byte]struct{}),
	}
	copy(ret.PathFragment, n.PathFragment)
	for k, v := range n.ChildCommitments {
		ret.ChildCommitments[k] = v.Clone()
	}
	for k, v := range n.modifiedChildren {
		ret.modifiedChildren[k] = v
	}
	return ret
}

func (n *Node) CommitsToTerminal() bool {
	return n.newTerminal != nil
}

func (n *Node) ChildKey(nodeKey kv.Key, childIndex byte) kv.Key {
	var buf bytes.Buffer
	buf.Write([]byte(nodeKey))
	buf.Write(n.PathFragment)
	buf.WriteByte(childIndex)
	return kv.Key(buf.Bytes())
}

func (n *Node) Write(w io.Writer) error {
	if err := util.WriteBytes16(w, n.PathFragment); err != nil {
		return err
	}

	var smallFlags byte
	if n.Terminal != nil {
		smallFlags = hasTerminalValueFlag
	}
	// compress children flags 32 bytes (if any)
	var flags [32]byte
	for i := range n.ChildCommitments {
		flags[i/8] |= 0x1 << (i % 8)
		smallFlags |= hasChildrenFlag
	}
	if err := util.WriteByte(w, smallFlags); err != nil {
		return err
	}
	// write terminal commitment if any
	if smallFlags&hasTerminalValueFlag != 0 {
		if err := n.Terminal.Write(w); err != nil {
			return err
		}
	}
	// write child commitments if any
	if smallFlags&hasChildrenFlag != 0 {
		if _, err := w.Write(flags[:]); err != nil {
			return err
		}
		for i := 0; i < 256; i++ {
			child, ok := n.ChildCommitments[uint8(i)]
			if !ok {
				continue
			}
			if err := child.Write(w); err != nil {
				return err
			}
		}
	}
	return nil
}

func (n *Node) Read(r io.Reader, setup CommitmentModel) error {
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
				n.ChildCommitments[ib] = setup.NewVectorCommitment()
				if err := n.ChildCommitments[ib].Read(r); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (n *Node) Bytes() []byte {
	var buf bytes.Buffer
	_ = n.Write(&buf)
	return buf.Bytes()
}
