package coretypes

import (
	"bytes"
	"encoding/binary"
	"io"
	"strconv"
)

type Uint16 uint16
type Uint32 uint32

func (i Uint16) Bytes() []byte {
	ret := make([]byte, 2)
	binary.LittleEndian.PutUint16(ret, (uint16)(i))
	return ret
}

func (i Uint16) String() string {
	return strconv.Itoa(int(i))
}

func NewUint16From2Bytes(data []byte) (Uint16, error) {
	if len(data) != 2 {
		return 0, ErrWrongDataConversion
	}
	return (Uint16)(binary.LittleEndian.Uint16(data)), nil
}

func NewUint32FromBytes(data []byte) (ret Uint32, err error) {
	err = ret.Read(bytes.NewReader(data))
	return
}

func (i Uint32) Bytes() []byte {
	ret := make([]byte, 4)
	binary.LittleEndian.PutUint32(ret, (uint32)(i))
	return ret
}

func (i Uint32) String() string {
	return strconv.Itoa((int)(i))
}

func (i *Uint32) Write(w io.Writer) error {
	_, err := w.Write(i.Bytes())
	return err
}

func (i *Uint32) Read(r io.Reader) error {
	var b [4]byte
	n, err := r.Read(b[:])
	if err != nil {
		return err
	}
	if n != 4 {
		return ErrWrongDataLength
	}
	t := binary.LittleEndian.Uint32(b[:])
	*i = (Uint32)(t)
	return nil
}
