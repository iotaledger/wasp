package util

import (
	"bytes"
	"io"

	"github.com/iotaledger/wasp/packages/hashing"
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

func GetHashValue(obj interface{ Bytes() []byte }) hashing.HashValue {
	return hashing.HashData(obj.Bytes())
}
