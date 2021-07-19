// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// package sb_utils implements Sandbox utility functions
package sandbox_utils //nolint:revive // TODO refactor to remove `_` from package name

import (
	"github.com/iotaledger/wasp/packages/iscp"
)

type utilImpl struct{}

func NewUtils() iscp.Utils {
	return utilImpl{}
}

func (u utilImpl) Base58() iscp.Base58 {
	return base58Util{}
}

func (u utilImpl) Hashing() iscp.Hashing {
	return hashUtil{}
}

func (u utilImpl) ED25519() iscp.ED25519 {
	return ed25519Util{}
}

func (u utilImpl) BLS() iscp.BLS {
	return blsUtil{}
}
