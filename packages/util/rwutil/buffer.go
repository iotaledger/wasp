// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package rwutil

import "io"

// Buffer implements a hyper-simple and efficient in-memory read/write stream.
// It will read from the start of the buffer and write to the end of the buffer.
type Buffer []byte

var _ io.ReadWriter = new(Buffer)

func (buf *Buffer) Read(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	if len(*buf) == 0 {
		return 0, io.EOF
	}
	n := copy(data, *buf)
	*buf = (*buf)[n:]
	return n, nil
}

func (buf *Buffer) Write(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	*buf = append(*buf, data...)
	return len(data), nil
}
