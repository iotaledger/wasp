// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package iscpcontract

import (
	_ "embed"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/iotaledger/wasp/packages/iscp"
)

// If you change ISCP.sol or ISCP.yul, you must recompile them.  You will need
// the `solc` binary installed in your system. Then, simply run `go generate`
// in this directory.

//go:generate solc --abi --bin-runtime --overwrite --revert-strings debug ISCP.sol -o .
var (
	// EVMAddress is the arbitrary address on which the standard
	// ISCP EVM contract lives
	EVMAddress = common.HexToAddress("0x1074")
	//go:embed ISCP.abi
	ABI string
	//go:embed ISCP.bin-runtime
	bytecodeHex string
)

//go:generate sh -c "solc --strict-assembly ISCP.yul | awk '/Binary representation:/ { getline; print $0 }' > ISCPYul.bin-runtime"
var (
	evmYulAddress = common.HexToAddress("0x1075")
	//go:embed ISCPYul.bin-runtime
	yulBytecodeHex string
)

func init() {
	if iscp.ChainIDLength != 20 {
		panic("ChainID length does not match bytes20 in ISCP.sol")
	}
}

// DeployOnGenesis sets up the initial state of the ISCP EVM contract
// which will go into the EVM genesis block
func DeployOnGenesis(genesisAlloc core.GenesisAlloc, chainID *iscp.ChainID) {
	// TODO: Execute a constructor instead of filling out storage manually
	// Note: To get the storage layout: solc --storage-layout ISCP.sol | tail -n +4 | jq .
	// slot 0: [offset 0 = ISCP.chainID]
	slot0 := common.Hash{}
	copy(slot0[:], chainID.Bytes())

	genesisAlloc[EVMAddress] = core.GenesisAccount{
		Code: common.FromHex(strings.TrimSpace(bytecodeHex)),
		Storage: map[common.Hash]common.Hash{
			common.HexToHash("00"): slot0,
		},
		Balance: &big.Int{},
	}

	genesisAlloc[evmYulAddress] = core.GenesisAccount{
		Code:    common.FromHex(strings.TrimSpace(yulBytecodeHex)),
		Balance: &big.Int{},
	}
}
