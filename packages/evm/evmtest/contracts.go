// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmtest

import (
	_ "embed"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

// If you change any of the .sol files, you must recompile them.  You will need
// the `solc` binary installed in your system. Then, simply run `go generate`
// in this directory.

//go:generate solc --abi --bin --overwrite Storage.sol -o .
var (
	//go:embed Storage.abi
	StorageContractABI string
	//go:embed Storage.bin
	storageContractBytecodeHex string
	StorageContractBytecode    = common.FromHex(strings.TrimSpace(storageContractBytecodeHex))
)

//go:generate solc --abi --bin --bin-runtime --overwrite ERC20Basic.sol -o .
var (
	//go:embed ERC20Basic.abi
	ERC20ContractABI string
	//go:embed ERC20Basic.bin
	erc20ContractBytecodeHex string
	ERC20ContractBytecode    = common.FromHex(strings.TrimSpace(erc20ContractBytecodeHex))
	//deployed bytecode and runtime bytecode are different, see: https://ethereum.stackexchange.com/questions/13086/whats-the-difference-between-solcs-bin-bytecode-versus-bin-runtime
	//go:embed ERC20Basic.bin-runtime
	ERC20ContractRuntimeBytecodeHex string
	ERC20ContractRuntimeBytecode    = common.FromHex(strings.TrimSpace(ERC20ContractRuntimeBytecodeHex))
)

//go:generate solc --abi --bin --overwrite EndlessLoop.sol -o .
var (
	//go:embed EndlessLoop.abi
	LoopContractABI string
	//go:embed EndlessLoop.bin
	loopContractBytecodeHex string
	LoopContractBytecode    = common.FromHex(strings.TrimSpace(loopContractBytecodeHex))
)

//go:generate sh -c "solc --abi --bin --overwrite @iscmagic=`realpath ../../vm/core/evm/iscmagic` ISCTest.sol -o . && rm ISC.*"
var (
	//go:embed ISCTest.abi
	ISCTestContractABI string
	//go:embed ISCTest.bin
	iscTestContractBytecodeHex string
	ISCTestContractBytecode    = common.FromHex(strings.TrimSpace(iscTestContractBytecodeHex))
)

//go:generate solc --abi --bin --overwrite Fibonacci.sol -o .
var (
	//go:embed Fibonacci.abi
	FibonacciContractABI string
	//go:embed Fibonacci.bin
	fibonacciContractBytecodeHex string
	FibonacciContractByteCode    = common.FromHex(strings.TrimSpace(fibonacciContractBytecodeHex))
)

//go:generate solc --abi --bin --overwrite GasTestMemory.sol -o .
var (
	//go:embed GasTestMemory.abi
	GasTestMemoryContractABI string
	//go:embed GasTestMemory.bin
	gasTestMemoryContractBytecodeHex string
	GasTestMemoryContractBytecode    = common.FromHex(strings.TrimSpace(gasTestMemoryContractBytecodeHex))
)

//go:generate solc --abi --bin --overwrite GasTestStorage.sol -o .
var (
	//go:embed GasTestStorage.abi
	GasTestStorageContractABI string
	//go:embed GasTestStorage.bin
	gasTestStorageContractBytecodeHex string
	GasTestStorageContractBytecode    = common.FromHex(strings.TrimSpace(gasTestStorageContractBytecodeHex))
)

//go:generate solc --abi --bin --overwrite GasTestExecutionTime.sol -o .
var (
	//go:embed GasTestExecutionTime.abi
	GasTestExecutionTimeContractABI string
	//go:embed GasTestExecutionTime.bin
	gasTestExecutionTimeContractBytecodeHex string
	GasTestExecutionTimeContractBytecode    = common.FromHex(strings.TrimSpace(gasTestExecutionTimeContractBytecodeHex))
)
