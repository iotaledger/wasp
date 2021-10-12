// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package iscptest

import (
	_ "embed"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

// If you change ISCPTest.sol, you must recompile.  You will need
// the `solc` binary installed in your system. Then, simply run `go generate`
// in this directory.

//go:generate sh -c "solc --abi --bin --overwrite @iscpcontract=`realpath ../iscpcontract` ISCPTest.sol -o . && rm ISCP.*"
var (
	//go:embed ISCPTest.abi
	ISCPTestContractABI string
	//go:embed ISCPTest.bin
	iscpTestContractBytecodeHex string
	ISCPTestContractBytecode    = common.FromHex(strings.TrimSpace(iscpTestContractBytecodeHex))
)
