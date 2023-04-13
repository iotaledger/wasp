// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmtypes

import (
	"github.com/iotaledger/hive.go/serializer/v2/marshalutil"
)

func readBytes(m *marshalutil.MarshalUtil) (b []byte, err error) {
	var n uint32
	if n, err = m.ReadUint32(); err != nil {
		return nil, err
	}
	return m.ReadBytes(int(n))
}

func writeBytes(m *marshalutil.MarshalUtil, b []byte) {
	m.WriteUint32(uint32(len(b)))
	m.WriteBytes(b)
}
