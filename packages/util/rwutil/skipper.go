// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package rwutil

import (
	"errors"
	"io"
)

type Skipper struct {
	w     io.Writer
	ww    *Writer
	count int
}

var _ io.ReadWriter = new(PushBack)

func (skip *Skipper) Read(data []byte) (n int, err error) {
	if skip.ww.w == skip.w {
		return 0, errors.New("invalid skipper read")
	}
	n = len(data)
	skip.count += n
	return n, nil
}

func (skip *Skipper) Write(data []byte) (n int, err error) {
	if skip.count == 0 {
		// exhausted skip count, switch back and write to normal stream
		skip.ww.w = skip.w
		return skip.w.Write(data)
	}

	n = len(data)
	if n > skip.count {
		// partial skip, switch back and report error
		n = skip.count
		skip.count = 0
		skip.ww.w = skip.w
		return n, errors.New("partial skip attempt")
	}

	// skip was fulfilled
	skip.count -= n
	return n, nil
}
