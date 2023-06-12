// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package rwutil

import (
	"bytes"
	"errors"
	"io"
)

type PushBack struct {
	r   io.Reader
	rr  *Reader
	buf *bytes.Buffer
}

var _ io.ReadWriter = new(PushBack)

func (push *PushBack) Read(data []byte) (int, error) {
	nBuf, err := push.buf.Read(data)
	if err != nil {
		if !errors.Is(err, io.EOF) {
			return nBuf, err
		}

		// exhausted buffer, switch back to normal stream and re-read
		push.rr.r = push.r
		return push.r.Read(data)
	}

	// read was fulfilled from buffer?
	if nBuf == len(data) {
		return nBuf, nil
	}

	// partial read from buffer, switch back to normal stream and read rest
	push.rr.r = push.r
	nStream, err := push.r.Read(data[nBuf:])

	if errors.Is(err, io.EOF) {
		// exhausted stream, report partial amount from buffer
		return nBuf, nil
	}

	// report total amount read from buffer and stream
	return nBuf + nStream, err
}

func (push *PushBack) Write(data []byte) (n int, err error) {
	if push.rr.r == push.r {
		return 0, errors.New("invalid pushback write")
	}
	return push.buf.Write(data)
}
