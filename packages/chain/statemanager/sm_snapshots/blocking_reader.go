package sm_snapshots

import (
	"errors"
	"io"
)

type blockingReader struct {
	r   io.Reader
	err error
}

func NewBlockingReader(r io.Reader) io.Reader {
	return &blockingReader{r: r, err: nil}
}

// Enforces following in addition to contract of io.Reader interface Read method
// to make it more convenient to use:
//   - If no errors occurred while reading, it always reads len(p) bytes and returns
//     no error. It may block to read all the required bytes. It does not "return
//     what is available" as is convention in io.Reader Read method.
//   - If error occurred while reading and len(p) bytes cannot be read, it returns
//     as many bytes as was read and no error. On another call it will return the
//     collected errors.
func (br *blockingReader) Read(p []byte) (int, error) {
	if br.err != nil {
		result := br.err
		br.err = nil
		return 0, result
	}
	read := 0
	for read < len(p) {
		r, e := br.r.Read(p[read:])
		br.err = errors.Join(br.err, e)
		if r == 0 {
			if read == 0 {
				result := br.err
				br.err = nil
				return read, result
			}
			return read, nil
		}
		read += r
	}
	return read, nil
}
