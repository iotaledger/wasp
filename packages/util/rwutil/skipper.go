// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package rwutil

import (
	"errors"
	"io"
)

// Skipper implements a skip wrapper for any writer stream.
// It works kind of the opposite of the PushBack wrapper.
// It allows you to dummy-read data from the stream and counts the bytes read.
// It will dummy-write these bytes first, and then resume writing to the wrapped stream
// The skip Reader is only valid for this Writer until it resumes the stream.
// See accounts.getNFTData() and accounts.saveNFTData() for an example
// of how Pushback and Skipper work in conjunction.
type Skipper struct {
	w     io.Writer
	ww    *Writer
	count int
}

var _ io.ReadWriter = new(PushBack)

func (skip *Skipper) Read(data []byte) (n int, err error) {
	// if we have an associated Writer make sure to prevent Skipper
	// reading after switching back to the wrapped writer
	if skip.ww != nil && skip.ww.w == skip.w {
		return 0, errors.New("invalid skipper read")
	}
	n = len(data)
	skip.count += n
	return n, nil
}

func (skip *Skipper) Write(data []byte) (nSkipped int, err error) {
	if skip.count == 0 {
		// exhausted skip count
		// if we have an associated Writer switch it back to the wrapped stream
		if skip.ww != nil {
			skip.ww.w = skip.w
		}
		// write to wrapped stream
		return skip.w.Write(data)
	}

	nSkipped = len(data)

	// skip was completely fulfilled?
	if nSkipped <= skip.count {
		skip.count -= nSkipped
		return nSkipped, nil
	}

	// partial skip
	nSkipped = skip.count
	skip.count = 0

	// if we have an associated Writer switch it back to the wrapped stream
	if skip.ww != nil {
		skip.ww.w = skip.w
	}

	// attempt to write the rest to the wrapped stream
	nStream, err := skip.w.Write(data[nSkipped:])

	// report total amount written
	return nSkipped + nStream, err
}
