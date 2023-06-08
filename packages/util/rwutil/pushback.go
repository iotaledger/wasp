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

func (pb *PushBack) Read(data []byte) (int, error) {
	nBuf, err := pb.buf.Read(data)
	if err != nil {
		if !errors.Is(err, io.EOF) {
			return nBuf, err
		}

		// exhausted buffer, switch back to normal stream and re-read
		pb.rr.r = pb.r
		return pb.r.Read(data)
	}

	// read was fulfilled from buffer?
	if nBuf == len(data) {
		return nBuf, nil
	}

	// partial read from buffer, switch back to normal stream and read rest
	pb.rr.r = pb.r
	nStream, err := pb.r.Read(data[nBuf:])

	if errors.Is(err, io.EOF) {
		// exhausted stream, report partial amount from buffer
		return nBuf, nil
	}

	// report total amount read from buffer and stream
	return nBuf + nStream, err
}

func (pb *PushBack) Write(data []byte) (n int, err error) {
	if pb.rr.r == pb.r {
		return 0, errors.New("invalid pushback write")
	}
	return pb.buf.Write(data)
}
