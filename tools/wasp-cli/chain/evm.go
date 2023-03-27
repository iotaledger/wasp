// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package chain

import (
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/packages/vm/core/evm"
)

type evmDeployParams struct {
	ChainID         uint16
	BlockKeepAmount int32
}

func (d *evmDeployParams) initFlags(cmd *cobra.Command) {
	cmd.Flags().Uint16VarP(&d.ChainID, "evm-chainid", "", evm.DefaultChainID, "ChainID")
	cmd.Flags().Int32VarP(&d.BlockKeepAmount, "evm-block-keep-amount", "", evm.BlockKeepAmountDefault, "Amount of blocks to keep in DB (-1: keep all blocks)")
}
