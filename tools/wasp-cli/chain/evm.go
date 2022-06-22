// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"encoding/base64"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/spf13/cobra"
)

type evmDeployParams struct {
	ChainID         uint16
	alloc           []string
	allocBase64     string
	GasRatio        util.Ratio32
	BlockGasLimit   uint64
	BlockKeepAmount int32
}

func (d *evmDeployParams) initFlags(cmd *cobra.Command) {
	cmd.Flags().Uint16VarP(&d.ChainID, "evm-chainid", "", evm.DefaultChainID, "ChainID")
	cmd.Flags().StringSliceVarP(&d.alloc, "evm-alloc", "", nil, "Genesis allocation (format: <address>:<wei>,<address>:<wei>,...)")
	cmd.Flags().StringVarP(&d.allocBase64, "evm-alloc-bytes", "", "", "Genesis allocation (base64-encoded)")
	d.GasRatio = util.Ratio32{A: 1, B: 1}
	cmd.Flags().VarP(&d.GasRatio, "evm-gas-ratio", "", "ISC Gas : EVM gas ratio")
	cmd.Flags().Uint64VarP(&d.BlockGasLimit, "evm-gas-limit", "", evm.BlockGasLimitDefault, "Block gas limit")
	cmd.Flags().Int32VarP(&d.BlockKeepAmount, "evm-block-keep-amount", "", evm.BlockKeepAmountDefault, "Amount of blocks to keep in DB (-1: keep all blocks)")
}

func (d *evmDeployParams) getGenesis(def core.GenesisAlloc) core.GenesisAlloc {
	if len(d.alloc) == 0 && d.allocBase64 == "" {
		return def
	}
	if len(d.alloc) != 0 && d.allocBase64 != "" {
		log.Fatalf("--alloc and --alloc-bytes are mutually exclusive")
	}
	if len(d.alloc) != 0 {
		// --alloc provided
		ret := core.GenesisAlloc{}
		for _, arg := range d.alloc {
			parts := strings.Split(arg, ":")
			addr := common.HexToAddress(parts[0])
			wei := big.NewInt(0)
			_, ok := wei.SetString(parts[1], 10)
			if !ok {
				log.Fatalf("cannot parse wei")
			}
			ret[addr] = core.GenesisAccount{Balance: wei}
		}
		return ret
	}
	// --alloc-bytes provided
	b, err := base64.StdEncoding.DecodeString(d.allocBase64)
	log.Check(err)
	ret, err := evmtypes.DecodeGenesisAlloc(b)
	log.Check(err)
	return ret
}
