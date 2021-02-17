// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package sandbox_utils

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
)

type hashUtil struct{}

func (u hashUtil) Blake2b(data []byte) hashing.HashValue {
	return hashing.HashDataBlake2b(data)
}

func (u hashUtil) Sha3(data []byte) hashing.HashValue {
	return hashing.HashSha3(data)
}

func (u hashUtil) Hname(s string) coretypes.Hname {
	return coretypes.Hn(s)
}
