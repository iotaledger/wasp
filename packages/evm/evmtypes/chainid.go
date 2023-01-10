// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmtypes

import (
	"errors"

	"github.com/iotaledger/wasp/packages/util"
)

func EncodeChainID(chainID uint16) []byte {
	return util.Uint16To2Bytes(chainID)
}

func DecodeChainID(b []byte, def ...uint16) (uint16, error) {
	if b == nil {
		if len(def) == 0 {
			return 0, errors.New("cannot decode nil bytes")
		}
		return def[0], nil
	}
	return util.Uint16From2Bytes(b)
}

func MustDecodeChainID(b []byte, def ...uint16) uint16 {
	r, err := DecodeChainID(b, def...)
	if err != nil {
		panic(err)
	}
	return r
}
