package trie

import (
	"bytes"
	"encoding/hex"
	"errors"
	"io"
)

// Tcommitment (short for terminal commitment) commits to data of arbitrary size.
type Tcommitment struct {
	data    []byte
	isValue bool
}

// tcommitment is encoded as [header | data]
// where header = isValue (1 bit) | data size (7 bits)
const (
	tcommitmentIsValueMask  = 0x80
	tcommitmentDataSizeMask = tcommitmentIsValueMask - 1

	tcommitmentMaxSizeBytes    = 64
	tcommitmentHeaderSizeBytes = 1

	// if len(value) > tcommitmentDataSizeMax, tcommitment data will
	// be hash(value) which is 20 bytes
	tcommitmentDataSizeMax = tcommitmentMaxSizeBytes - tcommitmentHeaderSizeBytes
)

func init() {
	assertf(tcommitmentDataSizeMax <= tcommitmentDataSizeMask, "tcommitmentDataSizeMax <= tcommitmentDataSizeMask")
}

func CommitToData(data []byte) *Tcommitment {
	if len(data) == 0 {
		// empty slice -> no data (deleted)
		return nil
	}
	var commitmentBytes []byte
	var isValue bool

	if len(data) > tcommitmentDataSizeMax {
		isValue = false
		// taking the hash as commitment data for long values
		hash := blake2b160(data)
		commitmentBytes = hash[:]
	} else {
		isValue = true
		// just cloning bytes. The data is its own commitment
		commitmentBytes = concat(data)
	}
	assertf(len(commitmentBytes) <= tcommitmentDataSizeMax,
		"len(commitmentBytes) <= m.tcommitmentDataSizeMax")
	return &Tcommitment{
		data:    commitmentBytes,
		isValue: isValue,
	}
}

func newTerminalCommitment() *Tcommitment {
	// all 0 non hashed value
	return &Tcommitment{
		data:    make([]byte, 0, HashSizeBytes),
		isValue: false,
	}
}

func (t *Tcommitment) Equals(o *Tcommitment) bool {
	return bytes.Equal(t.data, o.data)
}

func (t *Tcommitment) Clone() *Tcommitment {
	return &Tcommitment{
		data:    concat(t.data),
		isValue: t.isValue,
	}
}

func (t *Tcommitment) Write(w io.Writer) error {
	assertf(len(t.data) <= tcommitmentDataSizeMax, "size <= tcommitmentDataSizeMax")
	size := byte(len(t.data))
	if t.isValue {
		size |= tcommitmentIsValueMask
	}
	if err := writeByte(w, size); err != nil {
		return err
	}
	_, err := w.Write(t.data)
	return err
}

func (t *Tcommitment) Read(r io.Reader) error {
	var err error
	var l byte
	if l, err = readByte(r); err != nil {
		return err
	}
	t.isValue = (l & tcommitmentIsValueMask) != 0
	l &= tcommitmentDataSizeMask
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

func (t *Tcommitment) Bytes() []byte {
	return mustBytes(t)
}

func (t *Tcommitment) String() string {
	return hex.EncodeToString(t.data)
}

func (t *Tcommitment) ExtractValue() ([]byte, bool) {
	if t.isValue {
		return t.data, true
	}
	return nil, false
}
