// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package iscmagic

import (
	_ "embed"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"github.com/iotaledger/wasp/packages/kv/codec"
)

// If you change any of the .sol files, you must recompile.  You will need
// the `solc` binary installed in your system. Then, simply run `go generate`
// in this directory.

const (
	addressTypeISCMagic = iota
	addressTypeERC20BaseTokens
	addressTypeERC20NativeTokens
)

var (
	AddressPrefix = []byte{0x10, 0x74}
	Address       = makeMagicAddress(addressTypeISCMagic, nil)
)

//go:generate sh -c "solc --abi --overwrite @iscmagic=`realpath .` ISCSandbox.sol -o ."
var (
	//go:embed ISCSandbox.abi
	SandboxABI string
)

//go:generate sh -c "solc --abi --overwrite @iscmagic=`realpath .` ISCUtil.sol -o ."
var (
	//go:embed ISCUtil.abi
	UtilABI string
)

//go:generate sh -c "solc --abi --overwrite @iscmagic=`realpath .` ISCAccounts.sol -o ."
var (
	//go:embed ISCAccounts.abi
	AccountsABI string
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

//go:generate sh -c "solc --abi --bin-runtime --storage-layout --overwrite @iscmagic=`realpath .` ERC20NativeTokens.sol -o ."
var (
	//go:embed ERC20NativeTokens.abi
	ERC20NativeTokensABI string
	//go:embed ERC20NativeTokens.bin-runtime
	erc20NativeTokensRuntimeBytecodeHex string
	ERC20NativeTokensRuntimeBytecode    = common.FromHex(strings.TrimSpace(erc20NativeTokensRuntimeBytecodeHex))
)

func ERC20NativeTokensAddress(foundrySN uint32) common.Address {
	return makeMagicAddress(addressTypeERC20NativeTokens, codec.EncodeUint32(foundrySN))
}

func makeMagicAddress(kind byte, payload []byte) common.Address {
	var ret common.Address
	// 2 bytes 1074 prefix + 1 byte "kind"
	if len(payload) > common.AddressLength-3 {
		panic("makeMagicAddress: invalid payload length")
	}
	copy(ret[0:2], AddressPrefix)
	ret[2] = kind
	copy(ret[3:], payload)
	return ret
}

func ERC20NativeTokensFoundrySN(addr common.Address) uint32 {
	return codec.MustDecodeUint32(addr[3:7])
}
