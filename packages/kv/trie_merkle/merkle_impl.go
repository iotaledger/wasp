// Package trie_merkle implements trie.CommitmentModel based on the Merkle tree
package trie_merkle

import (
	"bytes"
	"encoding/hex"
	"io"

	"github.com/iotaledger/wasp/packages/kv/trie"
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/crypto/blake2b"
	"golang.org/x/xerrors"
)

// implements commitment scheme based on blake2b hashing

type terminalCommitment struct {
	bytes    [32]byte
	lenPlus1 uint8
}

type vectorCommitment [32]byte

// Model implements 256+ Trie based on Merkle tree, i.e. on hashing with blake2b
type commitmentModel struct{}

var (
	_     trie.VCommitment = &vectorCommitment{}
	_     trie.TCommitment = &terminalCommitment{}
	Model                  = &commitmentModel{} // singleton
)

func (m *commitmentModel) NewTerminalCommitment() trie.TCommitment {
	return &terminalCommitment{}
}

func (m *commitmentModel) NewVectorCommitment() trie.VCommitment {
	return &vectorCommitment{}
}

func (m *commitmentModel) CommitToNode(n *trie.Node) trie.VCommitment {
	var hashes [258]*[32]byte

	empty := true
	for i, c := range n.ChildCommitments {
		hashes[i] = (*[32]byte)(c.(*vectorCommitment))
		empty = false
	}
	if n.Terminal != nil {
		hashes[256] = &n.Terminal.(*terminalCommitment).bytes
		empty = false
	}
	if empty {
		return nil
	}
	tmp := commitToData(n.PathFragment)
	hashes[257] = &tmp
	ret := (vectorCommitment)(hashVector(&hashes))
	return &ret
}

func (m *commitmentModel) CommitToData(data []byte) trie.TCommitment {
	if len(data) == 0 {
		// empty slice -> no data (deleted)
		return nil
	}
	return commitToTerminal(data)
}

func (m *commitmentModel) VectorCommitmentFromBytes(data []byte) (trie.VCommitment, error) {
	ret := m.NewVectorCommitment()
	if err := ret.Read(bytes.NewReader(data)); err != nil {
		return nil, err
	}
	return ret, nil
}

func (v *vectorCommitment) Bytes() []byte {
	var buf bytes.Buffer
	_ = v.Write(&buf)
	return buf.Bytes()
}

func (v *vectorCommitment) Read(r io.Reader) error {
	_, err := r.Read((*v)[:])
	return err
}

func (v *vectorCommitment) Write(w io.Writer) error {
	_, err := w.Write((*v)[:])
	return err
}

func (v *vectorCommitment) String() string {
	return hex.EncodeToString(v[:])
}

func (v *vectorCommitment) Clone() trie.VCommitment {
	if v == nil {
		return nil
	}
	ret := *v
	return &ret
}

func (v *vectorCommitment) Update(delta trie.VCommitment) {
	m, ok := delta.(*vectorCommitment)
	if !ok {
		panic("hash commitment expected")
	}
	*v = *m
}

func (t *terminalCommitment) Write(w io.Writer) error {
	if err := util.WriteByte(w, t.lenPlus1); err != nil {
		return err
	}
	l := byte(32)
	if t.lenPlus1 > 0 {
		l = t.lenPlus1 - 1
	}
	_, err := w.Write(t.bytes[:l])
	return err
}

func (t *terminalCommitment) Read(r io.Reader) error {
	var err error
	if t.lenPlus1, err = util.ReadByte(r); err != nil {
		return err
	}
	if t.lenPlus1 > 33 {
		return xerrors.New("terminal size byte must be <= 33")
	}
	l := byte(32)
	if t.lenPlus1 > 0 {
		l = t.lenPlus1 - 1
	}
	t.bytes = [32]byte{}
	n, err := r.Read(t.bytes[:l])
	if err != nil {
		return err
	}
	if n != int(l) {
		return xerrors.New("bad data length")
	}
	return nil
}

func (t *terminalCommitment) Bytes() []byte {
	var buf bytes.Buffer
	_ = t.Write(&buf)
	return buf.Bytes()
}

func (t *terminalCommitment) String() string {
	return hex.EncodeToString(t.bytes[:])
}

func (t *terminalCommitment) Equal(another trie.CommitmentBase) bool {
	if t == nil && another == nil {
		return true
	}
	if t == nil || another == nil {
		return false
	}
	a, ok := another.(*terminalCommitment)
	if !ok {
		return false
	}
	return *t == *a
}

func (t *terminalCommitment) Clone() trie.TCommitment {
	if t == nil {
		return nil
	}
	ret := *t
	return &ret
}

// return value of the terminal commitment and a flag which indicates if it is a hashed value (true) or original data (false)
func (t *terminalCommitment) value() ([]byte, bool) {
	return t.bytes[:t.lenPlus1-1], t.lenPlus1 == 0
}

func hashVector(hashes *[258]*[32]byte) [32]byte {
	var buf [258 * 32]byte // 8 KB + 32 B + 32 B
	for i, h := range hashes {
		if h == nil {
			continue
		}
		pos := 32 * int(i)
		copy(buf[pos:pos+32], h[:])
	}
	return blake2b.Sum256(buf[:])
}

func commitToData(data []byte) (ret [32]byte) {
	if len(data) <= 32 {
		copy(ret[:], data)
	} else {
		ret = blake2b.Sum256(data)
	}
	return
}

func commitToTerminal(data []byte) *terminalCommitment {
	ret := &terminalCommitment{
		bytes: commitToData(data),
	}
	if len(data) <= 32 {
		ret.lenPlus1 = uint8(len(data)) + 1 // 1-33
	}
	return ret
}
