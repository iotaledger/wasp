// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package rwutil

import (
	"io"
)

type Counter struct {
	count uint32
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

func (c *Counter) Close() {
	c.count = 0
	if c.rr != nil {
		c.rr.r = c.r
		c.rr = nil
		return
	}
	if c.ww != nil {
		c.ww.w = c.w
		c.ww = nil
		return
	}
	panic("already closed")
}

func (c *Counter) Count() uint32 {
	return c.count
}

func (c *Counter) Read(buf []byte) (int, error) {
	bytes, err := c.r.Read(buf)
	c.count += uint32(bytes)
	return bytes, err
}

func (c *Counter) Write(buf []byte) (int, error) {
	bytes, err := c.w.Write(buf)
	c.count += uint32(bytes)
	return bytes, err
}
