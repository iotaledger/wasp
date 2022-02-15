package merkle_trie

import (
	"bytes"
	"github.com/iotaledger/wasp/packages/kv/trie"
	"github.com/iotaledger/wasp/packages/util"
	"golang.org/x/xerrors"
	"io"
)

type Proof struct {
	Key  []byte
	Path []*ProofElement
}

type ProofElement struct {
	PathFragment []byte
	Children     map[byte]*[32]byte
	Terminal     *[32]byte
	ChildIndex   int
}

func ProofFromBytes(data []byte) (*Proof, error) {
	ret := &Proof{}
	if err := ret.Read(bytes.NewReader(data)); err != nil {
		return nil, err
	}
	return ret, nil
}

// Proof converts generic proof path to the Merkle proof path
func (m *trieSetup) Proof(path *trie.ProofPath) *Proof {
	ret := &Proof{
		Key:  path.Key,
		Path: make([]*ProofElement, len(path.Path)),
	}
	for i, eg := range path.Path {
		em := &ProofElement{
			PathFragment: eg.Node.PathFragment,
			Children:     make(map[byte]*[32]byte),
			Terminal:     nil,
			ChildIndex:   eg.ChildIndex,
		}
		if eg.Node.Terminal != nil {
			em.Terminal = (*[32]byte)(eg.Node.Terminal.(*hashCommitment))
		}
		for k, v := range eg.Node.Children {
			if int(k) == eg.ChildIndex {
				// skipping the commitment which must come from the next child. 256 and 257 will be skipped too
				continue
			}
			em.Children[k] = (*[32]byte)(v.(*hashCommitment))
		}
		ret.Path[i] = em
	}
	return ret
}

// MustKeyTerminal returns key and terminal commitment the proof is about. If it returns (?, nil) it means it is proof of absence
// It does not verify the proof, so this function should be used only after Validate()
func (p *Proof) MustKeyTerminal() ([]byte, *[32]byte) {
	if len(p.Path) == 0 {
		return nil, nil
	}
	lastElem := p.Path[len(p.Path)-1]
	switch {
	case lastElem.ChildIndex < 256:
		if _, ok := lastElem.Children[byte(lastElem.ChildIndex)]; ok {
			panic("nil child commitment expected for proof of absence")
		}
		return p.Key, nil
	case lastElem.ChildIndex == 256:
		return p.Key, lastElem.Terminal
	case lastElem.ChildIndex == 257:
		return p.Key, nil
	}
	panic("wrong lastElem.ChildIndex")
}

func (p *Proof) MustIsProofOfAbsence() bool {
	_, r := p.MustKeyTerminal()
	return r == nil
}

func (p *Proof) Validate(root *[32]byte) error {
	if len(p.Path) == 0 {
		if root != nil {
			return xerrors.New("proof is empty")
		}
		return nil
	}
	c, err := p.verify(0, 0)
	if err != nil {
		return err
	}
	if c != *root {
		return xerrors.New("commitment not equal to the root")
	}
	return nil
}

func (p *Proof) verify(pathIdx, keyIdx int) ([32]byte, error) {
	assert(pathIdx < len(p.Path), "assertion: pathIdx < len(p.Path)")
	assert(keyIdx <= len(p.Key), "assertion: keyIdx <= len(p.Key)")

	elem := p.Path[pathIdx]
	tail := p.Key[keyIdx:]
	isPrefix := bytes.HasPrefix(tail, elem.PathFragment)
	last := pathIdx == len(p.Path)-1
	if !last && !isPrefix {
		return [32]byte{}, xerrors.Errorf("wrong proof: proof path does not follow the key. Path position: %d, key position %d", pathIdx, keyIdx)
	}
	if !last {
		assert(isPrefix, "assertion: isPrefix")
		if elem.ChildIndex > 255 {
			return [32]byte{}, xerrors.Errorf("wrong proof: wrong child index. Path position: %d, key position %d", pathIdx, keyIdx)
		}
		if _, ok := elem.Children[byte(elem.ChildIndex)]; ok {
			return [32]byte{}, xerrors.Errorf("wrong proof: unexpected commitment at child index %d. Path position: %d, key position %d", elem.ChildIndex, pathIdx, keyIdx)
		}
		nextKeyIdx := keyIdx + len(elem.PathFragment) + 1
		if nextKeyIdx > len(p.Key) {
			return [32]byte{}, xerrors.Errorf("wrong proof: proof path out of key bounds. Path position: %d, key position %d", pathIdx, keyIdx)
		}
		c, err := p.verify(pathIdx+1, nextKeyIdx)
		if err != nil {
			return [32]byte{}, err
		}
		return elem.hashIt(&c), nil
	}
	// it is the last in the path
	if elem.ChildIndex < 256 {
		c := elem.Children[byte(elem.ChildIndex)]
		if c != nil {
			return [32]byte{}, xerrors.Errorf("wrong proof: child commitment of the last element expected to be nil. Path position: %d, key position %d", pathIdx, keyIdx)
		}
		return elem.hashIt(nil), nil
	}
	if elem.ChildIndex != 256 && elem.ChildIndex != 257 {
		return [32]byte{}, xerrors.Errorf("wrong proof: child index expected to be 256 or 257. Path position: %d, key position %d", pathIdx, keyIdx)
	}
	return elem.hashIt(nil), nil
}

func (e *ProofElement) hashIt(missingCommitment *[32]byte) [32]byte {
	var hashes [258]*hashCommitment
	for idx, c := range e.Children {
		hashes[idx] = (*hashCommitment)(c)
	}
	hashes[256] = (*hashCommitment)(e.Terminal)
	hashes[257] = hashData(e.PathFragment)
	if e.ChildIndex < 256 {
		hashes[e.ChildIndex] = (*hashCommitment)(missingCommitment)
	}
	return *hashHashes(&hashes)
}

func assert(cond bool, err interface{}) {
	if !cond {
		panic(err)
	}
}

func (p *Proof) Write(w io.Writer) {
	_ = util.WriteBytes16(w, p.Key)
	_ = util.WriteUint16(w, uint16(len(p.Path)))
	for _, e := range p.Path {
		e.Write(w)
	}
}

func (p *Proof) Read(r io.Reader) error {
	var err error
	if p.Key, err = util.ReadBytes16(r); err != nil {
		return err
	}
	var size uint16
	if err = util.ReadUint16(r, &size); err != nil {
		return err
	}
	p.Path = make([]*ProofElement, size)
	for i := range p.Path {
		p.Path[i] = &ProofElement{}
		if err = p.Path[i].Read(r); err != nil {
			return err
		}
	}
	return nil
}

const (
	hasTerminalValueFlag = 0x01
	hasChildrenFlag      = 0x02
)

func (e *ProofElement) Write(w io.Writer) {
	_ = util.WriteBytes16(w, e.PathFragment)
	_ = util.WriteUint16(w, uint16(e.ChildIndex))
	var smallFlags byte
	if e.Terminal != nil {
		smallFlags = hasTerminalValueFlag
	}
	// compress children flags 32 bytes (if any)
	var flags [32]byte
	for i := range e.Children {
		flags[i/8] |= 0x1 << (i % 8)
		smallFlags |= hasChildrenFlag
	}
	_ = util.WriteByte(w, smallFlags)
	// write terminal commitment if any
	if smallFlags&hasTerminalValueFlag != 0 {
		_, _ = w.Write(e.Terminal[:])
	}
	// write child commitments if any
	if smallFlags&hasChildrenFlag != 0 {
		_, _ = w.Write(flags[:])
		for i := 0; i < 256; i++ {
			child, ok := e.Children[uint8(i)]
			if !ok {
				continue
			}
			_, _ = w.Write(child[:])
		}
	}
}

func (e *ProofElement) Read(r io.Reader) error {
	var err error
	if e.PathFragment, err = util.ReadBytes16(r); err != nil {
		return err
	}
	var idx uint16
	if err := util.ReadUint16(r, &idx); err != nil {
		return err
	}
	e.ChildIndex = int(idx)
	var smallFlags byte
	if smallFlags, err = util.ReadByte(r); err != nil {
		return err
	}
	if smallFlags&hasTerminalValueFlag != 0 {
		e.Terminal = &[32]byte{}
		n, err := r.Read(e.Terminal[:])
		if err != nil {
			return err
		}
		if n != 32 {
			return xerrors.New("32 bytes expected")
		}
	} else {
		e.Terminal = nil
	}
	e.Children = make(map[byte]*[32]byte)
	if smallFlags&hasChildrenFlag != 0 {
		var flags [32]byte
		if _, err := r.Read(flags[:]); err != nil {
			return err
		}
		for i := 0; i < 256; i++ {
			ib := uint8(i)
			if flags[i/8]&(0x1<<(i%8)) != 0 {
				e.Children[ib] = &[32]byte{}
				n, err := r.Read(e.Children[ib][:])
				if err != nil {
					return err
				}
				if n != 32 {
					return xerrors.New("32 bytes expected")
				}
			}
		}
	}
	return nil

}
