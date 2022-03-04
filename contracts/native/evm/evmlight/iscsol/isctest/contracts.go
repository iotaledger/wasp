// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package isctest

import (
	_ "embed"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

// If you change ISCTest.sol, you must recompile.  You will need
// the `solc` binary installed in your system. Then, simply run `go generate`
// in this directory.

//go:generate sh -c "solc --abi --bin --overwrite @isccontract=`realpath ../` ISCTest.sol -o . && rm ISC.*"
var (
	//go:embed ISCTest.abi
	ISCTestContractABI string
	//go:embed ISCTest.bin
	iscTestContractBytecodeHex string
	ISCTestContractBytecode    = common.FromHex(strings.TrimSpace(iscTestContractBytecodeHex))
)
