// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"encoding/base64"

	"github.com/ethereum/go-ethereum/core"
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

type evmDeployParams struct {
	ChainID         uint16
	allocBase64     string
	GasRatio        util.Ratio32
	BlockGasLimit   uint64
	BlockKeepAmount int32
}

func (d *evmDeployParams) initFlags(cmd *cobra.Command) {
	cmd.Flags().Uint16VarP(&d.ChainID, "evm-chainid", "", evm.DefaultChainID, "ChainID")
	cmd.Flags().StringVarP(&d.allocBase64, "evm-alloc", "", "", "Genesis allocation (base64-encoded)")
	d.GasRatio = util.Ratio32{A: 1, B: 1}
	cmd.Flags().VarP(&d.GasRatio, "evm-gas-ratio", "", "ISC Gas : EVM gas ratio")
	cmd.Flags().Uint64VarP(&d.BlockGasLimit, "evm-gas-limit", "", evm.BlockGasLimitDefault, "Block gas limit")
	cmd.Flags().Int32VarP(&d.BlockKeepAmount, "evm-block-keep-amount", "", evm.BlockKeepAmountDefault, "Amount of blocks to keep in DB (-1: keep all blocks)")
}

func (d *evmDeployParams) getGenesis(def core.GenesisAlloc) core.GenesisAlloc {
	if d.allocBase64 == "" {
		return def
	}
	b, err := base64.StdEncoding.DecodeString(d.allocBase64)
	log.Check(err)
	ret, err := evmtypes.DecodeGenesisAlloc(b)
	log.Check(err)
	return ret
}
