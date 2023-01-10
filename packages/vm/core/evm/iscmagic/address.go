// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package iscmagic

import (
	"bytes"
	_ "embed"
	"errors"

	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/wasp/packages/kv/codec"
)

type addressKind uint8

const (
	addressKindISCMagic = addressKind(iota)
	addressKindERC20BaseTokens
	addressKindERC20NativeTokens
	addressKindERC721NFTs
	addressKindInvalid
)

var (
	AddressPrefix = []byte{0x10, 0x74}
	Address       = packMagicAddress(addressKindISCMagic, nil)
)

func ERC20NativeTokensAddress(foundrySN uint32) common.Address {
	return packMagicAddress(addressKindERC20NativeTokens, codec.EncodeUint32(foundrySN))
}

func ERC20NativeTokensFoundrySN(addr common.Address) (uint32, error) {
	kind, payload, err := unpackMagicAddress(addr)
	if err != nil {
		return 0, err
	}
	if kind != addressKindERC20NativeTokens {
		return 0, errors.New("ERC20NativeTokensFoundrySN: invalid address kind")
	}
	if !allZero(payload[4:]) {
		return 0, errors.New("ERC20NativeTokensFoundrySN: invalid address format")
	}
	return codec.MustDecodeUint32(payload[0:4]), nil
}

func packMagicAddress(kind addressKind, payload []byte) common.Address {
	var ret common.Address
	copy(ret[0:2], AddressPrefix)
	ret[2] = byte(kind)
	if len(payload) > len(ret[3:]) {
		panic("packMagicAddress: invalid payload length")
	}
	copy(ret[3:], payload)
	return ret
}

func unpackMagicAddress(addr common.Address) (addressKind, []byte, error) {
	if !bytes.Equal(addr[0:2], AddressPrefix) {
		return 0, nil, errors.New("unpackMagicAddress: expected magic address prefix")
	}
	kind := addressKind(addr[2])
	if kind >= addressKindInvalid {
		return 0, nil, errors.New("unpackMagicAddress: unknown address kind")
	}
	payload := addr[3:]
	return kind, payload, nil
}

func allZero(s []byte) bool {
	for _, v := range s {
		if v != 0 {
			return false
		}
	}
	return true
}
