package trie

import (
	"bytes"
	"io"
)

func Bytes(o interface{ Write(w io.Writer) }) []byte {
	var buf bytes.Buffer
	o.Write(&buf)
	return buf.Bytes()
}

type byteCounter int

func (b *byteCounter) Write(p []byte) (n int, err error) {
	*b = byteCounter(int(*b) + len(p))
	return 0, nil
}

func Size(o interface{ Write(w io.Writer) }) int {
	var ret byteCounter
	o.Write(&ret)
	return int(ret)
}

type sliceWriter []byte

func (w sliceWriter) Write(p []byte) (int, error) {
	if len(p) > len(w) {
		panic("sliceWriter: data does not fit the target")
	}
	copy(w, p)
	return len(p), nil
}

func NewSliceWriter(buf []byte) io.Writer {
	return sliceWriter(buf)
}
