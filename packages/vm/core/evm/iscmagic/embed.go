// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package iscmagic

import (
	_ "embed"
	"encoding/binary"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
)

// If you change any of the .sol files, you must recompile.  You will need
// the `solc` binary installed in your system. Then, simply run `go generate`
// in this directory.

//go:generate sh -c "solc --abi --overwrite ISC.sol -o ."
var (
	//go:embed ISC.abi
	ABI string
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

var ERC20BaseTokensAddress = iscAddressPlusOne()

func iscAddressPlusOne() common.Address {
	n := binary.BigEndian.Uint64(vm.ISCAddress.Bytes()[common.AddressLength-8:]) + 1
	var ret common.Address
	binary.BigEndian.PutUint64(ret[common.AddressLength-8:], n)
	return ret
}
