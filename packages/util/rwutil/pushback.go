// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package rwutil

import (
	"errors"
	"io"
)

// PushBack implements a pushback wrapper for any read stream.
// It uses an in-memory buffer that allows you to write data back to the stream.
// It will read this data first, and then resume reading from the wrapped stream.
// The pushback Writer is only valid for this Reader until it resumes the stream.
// See accounts.getNFTData() and accounts.saveNFTData() for an example
// of how Pushback and Skipper work in conjunction.
type PushBack struct {
	r   io.Reader
	rr  *Reader
	buf Buffer
}

var _ io.ReadWriter = new(PushBack)

func (push *PushBack) Read(data []byte) (int, error) {
	if len(push.buf) == 0 {
		// exhausted read buffer
		// if we have an associated Reader switch it back to the wrapped stream
		if push.rr != nil {
			push.rr.r = push.r
		}
		// read from wrapped stream
		return push.r.Read(data)
	}

	nPushed, err := push.buf.Read(data)
	if err != nil {
		return nPushed, err
	}

	// read was completely fulfilled from buffer?
	if nPushed == len(data) {
		return nPushed, nil
	}

	// partial read from buffer
	// if we have an associated Reader switch it back to the wrapped stream
	if push.rr != nil {
		push.rr.r = push.r
	}

	// attempt to read the rest from the wrapped stream
	nStream, err := push.r.Read(data[nPushed:])

	// special case, we don't return EOF here because we already read some bytes
	if errors.Is(err, io.EOF) {
		// exhausted stream, report partial amount from buffer
		return nPushed, nil
	}

	// report total amount read
	return nPushed + nStream, err
}

func (push *PushBack) Write(data []byte) (int, error) {
	// if we have an associated Reader make sure to prevent PushBack
	// writing after switching back to the wrapped reader
	if push.rr != nil && push.rr.r == push.r {
		panic("invalid pushback write")
	}
	return push.buf.Write(data)
}
