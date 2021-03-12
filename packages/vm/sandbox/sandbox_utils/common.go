// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// package sb_utils implements Sandbox utility functions
package sandbox_utils

import (
	"github.com/iotaledger/wasp/packages/coretypes"
)

type utilImpl struct {
}

func NewUtils() coretypes.Utils {
	return utilImpl{}
}

func (u utilImpl) Base58() coretypes.Base58 {
	return base58Util{}
}

func (u utilImpl) Hashing() coretypes.Hashing {
	return hashUtil{}
}

func (u utilImpl) ED25519() coretypes.ED25519 {
	return ed25519Util{}
}

func (u utilImpl) BLS() coretypes.BLS {
	return blsUtil{}
}
