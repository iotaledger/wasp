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
