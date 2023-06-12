// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package rwutil

import (
	"io"

	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
)

type MuReader struct {
	mu *marshalutil.MarshalUtil
}

var _ io.Reader = new(MuReader)

func (mu *MuReader) Read(buf []byte) (int, error) {
	bytes, err := mu.mu.ReadBytes(len(buf))
	copy(buf, bytes)
	return len(bytes), err
}
