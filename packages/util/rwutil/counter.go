// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package rwutil

import (
	"io"
)

// Counter implements a read/write stream that can wrap another stream.
// It will count the total number of bytes read/written from/to the wrapped stream.
type Counter struct {
	count int
	r     io.Reader
	rr    *Reader
	w     io.Writer
	ww    *Writer
}

var _ io.ReadWriter = new(Counter)

func NewReadCounter(rr *Reader) (ret *Counter) {
	ret = &Counter{r: rr.r, rr: rr}
	rr.r = ret
	return ret
}

func NewWriteCounter(ww *Writer) (ret *Counter) {
	ret = &Counter{w: ww.w, ww: ww}
	ww.w = ret
	return ret
}

func (counter *Counter) Close() {
	counter.count = 0
	if counter.rr != nil {
		counter.rr.r = counter.r
		counter.rr = nil
		return
	}
	if counter.ww != nil {
		counter.ww.w = counter.w
		counter.ww = nil
		return
	}
	panic("already closed")
}

func (counter *Counter) Count() int {
	return counter.count
}

func (counter *Counter) Read(data []byte) (int, error) {
	bytes, err := counter.r.Read(data)
	counter.count += bytes
	return bytes, err
}

func (counter *Counter) Write(data []byte) (int, error) {
	bytes, err := counter.w.Write(data)
	counter.count += bytes
	return bytes, err
}
