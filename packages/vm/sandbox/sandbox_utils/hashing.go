// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package sandbox_utils //nolint:revive // TODO refactor to remove `_` from package name

import (
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
)

type hashUtil struct{}

func (u hashUtil) Blake2b(data []byte) hashing.HashValue {
	return hashing.HashDataBlake2b(data)
}

func (u hashUtil) Sha3(data []byte) hashing.HashValue {
	return hashing.HashSha3(data)
}

func (u hashUtil) Hname(s string) iscp.Hname {
	return iscp.Hn(s)
}
