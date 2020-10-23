package coretypes

import (
	"bytes"
	"encoding/binary"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/mr-tron/base58"
	"io"
)

const EntryPointCodeLength = 4

type EntryPointCode [EntryPointCodeLength]byte

func NewEntryPointCodeFromBytes(data []byte) (ret EntryPointCode, err error) {
	err = ret.Read(bytes.NewReader(data))
	return
}

// NewEntryPointCodeFromFunctionName beware collisions!
// must always be checked agains the whole table for collisions and adjusted
func NewEntryPointCodeFromFunctionName(funname string) (ret EntryPointCode) {
	h := hashing.HashStrings(funname)
	copy(ret[:], h[:4])
	return
}

func NewEntryPointCodeFromUint32(n uint32) (ret EntryPointCode) {
	binary.LittleEndian.PutUint32(ret[:], n)
	return
}

func NewEntryPointCodeFromBase58(b58 string) (ret EntryPointCode, err error) {
	data, err := base58.Decode(b58)
	if err != nil {
		return
	}
	return NewEntryPointCodeFromBytes(data)
}

func (ec EntryPointCode) Uint32() uint32 {
	return binary.LittleEndian.Uint32(ec[:])

}

func (ec EntryPointCode) String() string {
	return base58.Encode(ec[:])
}

func (ec *EntryPointCode) Write(w io.Writer) error {
	_, err := w.Write(ec[:])
	return err
}

func (ec *EntryPointCode) Read(r io.Reader) error {
	n, err := r.Read(ec[:])
	if err != nil {
		return err
	}
	if n != EntryPointCodeLength {
		return ErrWrongDataLength
	}
	return nil
}
