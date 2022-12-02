package trie

import (
	"bytes"
	"encoding/hex"
	"errors"
	"io"
)

// blake2b_* contains implementation of commitment model for the trie
// based on `blake2b` 20 byte (160 bit) hashing.

// terminalCommitment commits to data of arbitrary size.
type terminalCommitment struct {
	data    []byte
	isValue bool
}

const (
	HashSizeBits  = 160
	HashSizeBytes = HashSizeBits / 8

	vectorLength                 = NumChildren + 2 // 16 children + terminal + path extension
	terminalCommitmentIndex      = NumChildren
	pathExtensionCommitmentIndex = NumChildren + 1
)

type Hash [HashSizeBytes]byte

// vectorCommitment is a blake2b hash of the vector elements
type vectorCommitment Hash

// terminalCommitment is encoded as [header | data]
// where header = isValue (1 bit) | data size (7 bits)
const (
	tCommitmentIsValueMask  = 0x80
	tCommitmentDataSizeMask = tCommitmentIsValueMask - 1

	tCommitmentMaxSizeBytes    = 64
	tCommitmentHeaderSizeBytes = 1

	// if len(value) > tCommitmentDataSizeMax, terminalCommitment data will
	// be hash(value) which is 20 bytes
	tCommitmentDataSizeMax = tCommitmentMaxSizeBytes - tCommitmentHeaderSizeBytes
)

func init() {
	assert(tCommitmentDataSizeMax <= tCommitmentDataSizeMask, "tCommitmentDataSizeMax <= tCommitmentDataSizeMask")
}

// updateNodeCommitment computes update to the node data and, optionally, updates existing commitment.
func updateNodeCommitment(mutate *nodeData, childUpdates map[byte]VCommitment, newTerminalUpdate TCommitment, pathExtension []byte) {
	for i, upd := range childUpdates {
		mutate.children[i] = upd
	}
	mutate.terminal = newTerminalUpdate // for hash commitment just replace
	mutate.pathExtension = pathExtension
	if mutate.ChildrenCount() == 0 && mutate.terminal == nil {
		return
	}
	v := vectorCommitment(makeHashVector(mutate).Hash())
	mutate.commitment = &v
}

// compressToHashSize hashes data if longer than hash size, otherwise copies it
func compressToHashSize(data []byte) (ret []byte, valueInCommitment bool) {
	if len(data) <= HashSizeBytes {
		ret = make([]byte, len(data))
		valueInCommitment = true
		copy(ret, data)
	} else {
		hash := blake2b160(data)
		ret = hash[:]
	}
	return
}

func CommitToData(data []byte) TCommitment {
	if len(data) == 0 {
		// empty slice -> no data (deleted)
		return nil
	}
	var commitmentBytes []byte
	var isValue bool

	if len(data) > tCommitmentDataSizeMax {
		isValue = false
		// taking the hash as commitment data for long values
		hash := blake2b160(data)
		commitmentBytes = hash[:]
	} else {
		isValue = true
		// just cloning bytes. The data is its own commitment
		commitmentBytes = concat(data)
	}
	assert(len(commitmentBytes) <= tCommitmentDataSizeMax,
		"len(commitmentBytes) <= m.tCommitmentDataSizeMax")
	return &terminalCommitment{
		data:    commitmentBytes,
		isValue: isValue,
	}
}

type hashVector [vectorLength][]byte

// makeHashVector makes the node vector to be hashed. Missing children are nil
func makeHashVector(nodeData *nodeData) *hashVector {
	hashes := &hashVector{}
	for i, c := range nodeData.children {
		if c != nil {
			hash := c.Hash()
			hashes[i] = hash[:]
		}
	}
	if nodeData.terminal != nil {
		// squeeze terminal it into the hash size, if longer than hash size
		hashes[terminalCommitmentIndex], _ = compressToHashSize(nodeData.terminal.Bytes())
	}
	pathExtensionCommitmentBytes, _ := compressToHashSize(nodeData.pathExtension)
	hashes[pathExtensionCommitmentIndex] = pathExtensionCommitmentBytes
	return hashes
}

func (hashes *hashVector) Hash() Hash {
	buf := make([]byte, vectorLength*HashSizeBytes)
	for i, h := range hashes {
		if h == nil {
			continue
		}
		pos := i * HashSizeBytes
		copy(buf[pos:pos+HashSizeBytes], h[:])
	}
	return blake2b160(buf)
}

// *vectorCommitment implements trie_go.VCommitment
var _ VCommitment = &vectorCommitment{}

func newVectorCommitment() *vectorCommitment {
	return &vectorCommitment{}
}

func (v *vectorCommitment) Clone() VCommitment {
	c := vectorCommitment{}
	copy(c[:], v[:])
	return &c
}

func (v *vectorCommitment) Hash() Hash {
	return Hash(*v)
}

func (v *vectorCommitment) Bytes() []byte {
	return mustBytes(v)
}

func (v *vectorCommitment) Read(r io.Reader) error {
	_, err := r.Read(v[:])
	return err
}

func (v *vectorCommitment) Write(w io.Writer) error {
	_, err := w.Write(v[:])
	return err
}

func (v *vectorCommitment) String() string {
	return hex.EncodeToString(v[:])
}

func (v *vectorCommitment) Equals(o VCommitment) bool {
	v2 := o.(*vectorCommitment)
	return *v == *v2
}

// *terminalCommitment implements trie_go.TCommitment
var _ TCommitment = &terminalCommitment{}

func newTerminalCommitment() *terminalCommitment {
	// all 0 non hashed value
	return &terminalCommitment{
		data:    make([]byte, 0, HashSizeBytes),
		isValue: false,
	}
}

func (t *terminalCommitment) Equals(o TCommitment) bool {
	t2 := o.(*terminalCommitment)
	return bytes.Equal(t.data, t2.data)
}

func (t *terminalCommitment) Clone() TCommitment {
	return &terminalCommitment{
		data:    concat(t.data),
		isValue: t.isValue,
	}
}

func (t *terminalCommitment) Write(w io.Writer) error {
	assert(len(t.data) <= tCommitmentDataSizeMax, "size <= tCommitmentDataSizeMax")
	size := byte(len(t.data))
	if t.isValue {
		size |= tCommitmentIsValueMask
	}
	if err := writeByte(w, size); err != nil {
		return err
	}
	_, err := w.Write(t.data)
	return err
}

func (t *terminalCommitment) Read(r io.Reader) error {
	var err error
	var l byte
	if l, err = readByte(r); err != nil {
		return err
	}
	t.isValue = (l & tCommitmentIsValueMask) != 0
	l &= tCommitmentDataSizeMask
	if l > 0 {
		t.data = make([]byte, l)

		n, err := r.Read(t.data)
		if err != nil {
			return err
		}
		if n != int(l) {
			return errors.New("bad data length")
		}
	}
	return nil
}

func (t *terminalCommitment) Bytes() []byte {
	return mustBytes(t)
}

func (t *terminalCommitment) String() string {
	return hex.EncodeToString(t.data[:])
}

func (t *terminalCommitment) ExtractValue() ([]byte, bool) {
	if t.isValue {
		return t.data, true
	}
	return nil, false
}

func ReadVectorCommitment(r io.Reader) (VCommitment, error) {
	ret := newVectorCommitment()
	if err := ret.Read(r); err != nil {
		return nil, err
	}
	return ret, nil
}

func ReadTerminalCommitment(r io.Reader) (TCommitment, error) {
	ret := newTerminalCommitment()
	if err := ret.Read(r); err != nil {
		return nil, err
	}
	return ret, nil
}

func VectorCommitmentFromBytes(data []byte) (VCommitment, error) {
	rdr := bytes.NewReader(data)
	ret, err := ReadVectorCommitment(rdr)
	if err != nil {
		return nil, err
	}
	if rdr.Len() > 0 {
		return nil, ErrNotAllBytesConsumed
	}
	return ret, nil
}

func TerminalCommitmentFromBytes(data []byte) (TCommitment, error) {
	rdr := bytes.NewReader(data)
	ret, err := ReadTerminalCommitment(rdr)
	if err != nil {
		return nil, err
	}
	if rdr.Len() > 0 {
		return nil, ErrNotAllBytesConsumed
	}
	return ret, nil
}
