package trie

import (
	"bytes"
	"io"
)

func Bytes(o interface{ Write(w io.Writer) error }) ([]byte, error) {
	var buf bytes.Buffer
	if err := o.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func MustBytes(o interface{ Write(w io.Writer) error }) []byte {
	ret, err := Bytes(o)
	if err != nil {
		panic(err)
	}
	return ret
}

type byteCounter int

func (b *byteCounter) Write(p []byte) (n int, err error) {
	*b = byteCounter(int(*b) + len(p))
	return 0, nil
}

func Size(o interface{ Write(w io.Writer) error }) (int, error) {
	var ret byteCounter
	if err := o.Write(&ret); err != nil {
		return 0, err
	}
	return int(ret), nil
}

func MustSize(o interface{ Write(w io.Writer) error }) int {
	ret, err := Size(o)
	if err != nil {
		panic(err)
	}
	return ret
}
