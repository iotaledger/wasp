package snapshot

import (
	"errors"
	"github.com/iotaledger/wasp/packages/util"
	"io"
	"os"
)

// Writer represents a specific format of the snapshot. It is a sequence of key/value pairs.
// It does not impose a specific order of key/value pairs
type Writer interface {
	WriteKeyValue(key, value []byte) error
	Stats() (int, int)
}

type Iterator interface {
	Iterate(func(k, v []byte) bool) error
}

type binKVDump struct {
	w         io.Writer
	kvCount   int
	byteCount int
}

func NewBinaryKVDump(w io.Writer) binKVDump {
	return binKVDump{w: w}
}

// binKVDump implements snapshot.Writer interface
var _ Writer = binKVDump{}

func (b binKVDump) WriteKeyValue(key, value []byte) error {
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

func (b binKVDump) Stats() (int, int) {
	return b.kvCount, b.byteCount
}

type binFileKVDump struct {
	binKVDump
	File *os.File
}

func CreateSnapshotFile(fname string) (*binFileKVDump, error) {
	file, err := os.Create(fname)
	if err != nil {
		return nil, err
	}
	return &binFileKVDump{
		binKVDump: NewBinaryKVDump(file),
		File:      file,
	}, nil
}

type binKVIterator struct {
	r         io.Reader
	kvCount   int
	byteCount int
}

func NewKVIterator(r io.Reader) binKVIterator {
	return binKVIterator{r: r}
}

type binSnapshotIterator struct {
	binKVIterator
	File *os.File
}

func (b binKVIterator) Iterate(fun func(k []byte, v []byte) bool) error {
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

func OpenSnapshotFile(fname string) (*binSnapshotIterator, error) {
	file, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	return &binSnapshotIterator{
		binKVIterator: NewKVIterator(file),
		File:          file,
	}, nil
}
