// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package rwutil

import (
	"errors"
	"io"
)

// Must will wrap a reader stream and will panic whenever an error occurs on that stream.
type Must struct {
	r io.Reader
}

var _ io.Reader = new(Must)

func (must *Must) Read(data []byte) (int, error) {
	bytes, err := must.r.Read(data)
	if err != nil && !errors.Is(err, io.EOF) {
		panic(err)
	}
	return bytes, err
}
