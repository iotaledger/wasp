package sm_snapshots

import (
	"io"
)

type readerWithClose struct {
	reader io.Reader
	close  func() error
}

var (
	_ io.Reader     = &readerWithClose{}
	_ io.Closer     = &readerWithClose{}
	_ io.ReadCloser = &readerWithClose{}
)

func NewReaderWithClose(r io.Reader, closeFun func() error) io.ReadCloser {
	return &readerWithClose{
		reader: r,
		close:  closeFun,
	}
}

func (r *readerWithClose) Read(b []byte) (int, error) {
	return r.reader.Read(b)
}

func (r *readerWithClose) Close() error {
	return r.close()
}
