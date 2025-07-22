// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package cmtlog

import (
	"github.com/iotaledger/wasp/v2/packages/kv/codec"
)

// LogIndex starts from 1. 0 is used as a nil value.
type LogIndex uint32

func (li LogIndex) AsUint32() uint32 {
	return uint32(li)
}

func (li LogIndex) Bytes() []byte {
	return codec.Encode[uint32](li.AsUint32())
}

func (li LogIndex) IsNil() bool {
	return li == 0
}

func (li LogIndex) Next() LogIndex {
	return LogIndex(li.AsUint32() + 1)
}

func (li LogIndex) Prev() LogIndex {
	if li == 0 {
		return li
	}
	return LogIndex(li.AsUint32() - 1)
}

func (li LogIndex) Sub(x uint32) LogIndex {
	if li.AsUint32() <= x {
		return NilLogIndex()
	}
	return LogIndex(li.AsUint32() - x)
}

func NilLogIndex() LogIndex {
	return LogIndex(0)
}

func MaxLogIndex(lis ...LogIndex) LogIndex {
	maxLI := NilLogIndex()
	for _, li := range lis {
		if li > maxLI {
			maxLI = li
		}
	}
	return maxLI
}
