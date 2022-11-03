// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package iscmagic

import (
	_ "embed"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

// If you change any of the .sol files, you must recompile.  You will need
// the `solc` binary installed in your system. Then, simply run `go generate`
// in this directory.

const (
	addressTypeISCMagic = iota
	addressTypeERC20BaseTokens
)

//go:generate sh -c "solc --abi --overwrite ISC.sol -o ."
var (
	//go:embed ISC.abi
	ABI string

	AddressPrefix = []byte{0x10, 0x74}
	Address       = makeMagicAddress(addressTypeISCMagic, nil)
)

//go:generate sh -c "solc --abi --overwrite @iscmagic=`realpath .` ISCPrivileged.sol -o ."
var (
	//go:embed ISCPrivileged.abi
	PrivilegedABI string
)

//go:generate sh -c "solc --abi --bin-runtime --overwrite @iscmagic=`realpath .` ERC20BaseTokens.sol -o ."
var (
	//go:embed ERC20BaseTokens.abi
	ERC20BaseTokensABI string
	//go:embed ERC20BaseTokens.bin-runtime
	erc20BaseRuntimeBytecodeHex    string
	ERC20BaseTokensRuntimeBytecode = common.FromHex(strings.TrimSpace(erc20BaseRuntimeBytecodeHex))
)

var ERC20BaseTokensAddress = makeMagicAddress(addressTypeERC20BaseTokens, nil)

func makeMagicAddress(kind byte, data []byte) common.Address {
	var ret common.Address
	if len(data) > len(ret)-3 {
		panic("makeMagicAddress: invalid data length")
	}
	copy(ret[0:2], AddressPrefix)
	ret[2] = kind
	copy(ret[3:], data)
	return ret
}
