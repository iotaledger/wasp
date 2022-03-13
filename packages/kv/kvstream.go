package kv

import (
	"errors"
	"github.com/iotaledger/wasp/packages/util"
	"io"
	"os"
)

// Interfaces for writing/reading persistent streams of key/values

// StreamWriter represents an interface specific to write a sequence of key/value pairs
type StreamWriter interface {
	Write(key, value []byte) error
	Stats() (int, int) // num k/v pairs and num bytes so far
}

// StreamIterator is an interface to iterate stream
type StreamIterator interface {
	Iterate(func(k, v []byte) bool) error
}

// binStreamWriter writes stream of k/v pairs in binary format.
// Keys encoding is 'bytes16' and values is 'bytes32'
type binStreamWriter struct {
	w         io.Writer
	kvCount   int
	byteCount int
}

func NewBinaryStreamWriter(w io.Writer) binStreamWriter {
	return binStreamWriter{w: w}
}

// binStreamWriter implements StreamWriter interface
var _ StreamWriter = &binStreamWriter{}

func (b *binStreamWriter) Write(key, value []byte) error {
	if err := util.WriteBytes16(b.w, key); err != nil {
		return err
	}
	b.byteCount += len(key) + 2
	if err := util.WriteBytes32(b.w, value); err != nil {
		return err
	}
	b.byteCount += len(value) + 4
	b.kvCount++
	return nil
}

func (b *binStreamWriter) Stats() (int, int) {
	return b.kvCount, b.byteCount
}

type binStreamIterator struct {
	r io.Reader
}

func NewBinaryStreamIterator(r io.Reader) binStreamIterator {
	return binStreamIterator{r: r}
}

func (b binStreamIterator) Iterate(fun func(k []byte, v []byte) bool) error {
	for {
		k, err := util.ReadBytes16(b.r)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		v, err := util.ReadBytes32(b.r)
		if err != nil {
			return err
		}
		if !fun(k, v) {
			return nil
		}
	}
}

type binStreamFileWriter struct {
	binStreamWriter
	File *os.File
}

// CreateKVStreamFile create a new k/v file
func CreateKVStreamFile(fname string) (*binStreamFileWriter, error) {
	file, err := os.Create(fname)
	if err != nil {
		return nil, err
	}
	return &binStreamFileWriter{
		binStreamWriter: NewBinaryStreamWriter(file),
		File:            file,
	}, nil
}

type binStreamFileIterator struct {
	binStreamIterator
	File *os.File
}

// OpenKVStreamFile opens existing file with k/v stream
func OpenKVStreamFile(fname string) (*binStreamFileIterator, error) {
	file, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	return &binStreamFileIterator{
		binStreamIterator: NewBinaryStreamIterator(file),
		File:              file,
	}, nil
}
