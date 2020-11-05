package coretypes

import (
	"bytes"
	"encoding/binary"
	"github.com/iotaledger/wasp/packages/hashing"
	"io"
	"strconv"
)

type EntryPointCode uint32

const FuncInit = "init"

var EntryPointCodeInit = NewEntryPointCodeFromFunctionName(FuncInit)

func NewEntryPointCodeFromBytes(data []byte) (ret EntryPointCode, err error) {
	err = ret.Read(bytes.NewReader(data))
	return
}

// NewEntryPointCodeFromFunctionName beware collisions: hash is only 4 bytes!
// must always be checked against the whole table for collisions and adjusted
func NewEntryPointCodeFromFunctionName(funname string) (ret EntryPointCode) {
	_ = ret.Read(bytes.NewReader(hashing.HashStrings(funname)[:4]))
	return ret
}

func (i EntryPointCode) Bytes() []byte {
	ret := make([]byte, 4)
	binary.LittleEndian.PutUint32(ret, (uint32)(i))
	return ret
}

func (i EntryPointCode) String() string {
	return strconv.Itoa((int)(i))
}

func (i *EntryPointCode) Write(w io.Writer) error {
	_, err := w.Write(i.Bytes())
	return err
}

func (i *EntryPointCode) Read(r io.Reader) error {
	var b [4]byte
	n, err := r.Read(b[:])
	if err != nil {
		return err
	}
	if n != 4 {
		return ErrWrongDataLength
	}
	t := binary.LittleEndian.Uint32(b[:])
	*i = (EntryPointCode)(t)
	return nil
}
