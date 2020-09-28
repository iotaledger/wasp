package util

import (
	"bytes"
	"errors"
	"github.com/iotaledger/wasp/packages/hashing"
	"io"
)

func Bytes(obj interface{ Write(io.Writer) error }) ([]byte, error) {
	var buf bytes.Buffer
	if err := obj.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func MustBytes(obj interface{ Write(io.Writer) error }) []byte {
	ret, err := Bytes(obj)
	if err != nil {
		panic(err)
	}
	return ret
}

func GetHashValue(obj interface{ Write(io.Writer) error }) hashing.HashValue {
	return *hashing.HashData(MustBytes(obj))
}

func WriteSequence16(w io.Writer, num int, elemFun func(i int) interface{ Write(io.Writer) error }) error {
	if num > MaxUint16 {
		return errors.New("WriteSequence16: too long slice")
	}
	if err := WriteUint16(w, uint16(num)); err != nil {
		return err
	}
	for i := 0; i < num; i++ {
		if err := elemFun(i).Write(w); err != nil {
			return err
		}
	}
	return nil
}
