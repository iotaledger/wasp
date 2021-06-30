// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmcli

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/iotaledger/wasp/contracts/native/evmchain"
	"github.com/iotaledger/wasp/packages/evm/evmtest"
	"github.com/spf13/cobra"
)

type DeployParams struct {
	genesis    []string
	GasPerIOTA uint64
}

func (d *DeployParams) InitFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringSliceVarP(&d.genesis, "genesis", "s", nil, "genesis allocation (format: <address>:<wei>,...)")
	cmd.PersistentFlags().Uint64VarP(&d.GasPerIOTA, "gas-per-iota", "f", evmchain.DefaultGasPerIota, "Gas per IOTA charged as fee")
}

func (d *DeployParams) GetGenesis() core.GenesisAlloc {
	if len(d.genesis) == 0 {
		return core.GenesisAlloc{
			evmtest.FaucetAddress: {Balance: evmtest.FaucetSupply},
		}
	}
	ret := core.GenesisAlloc{}
	for _, s := range d.genesis {
		parts := strings.Split(s, ":")
		addr := common.HexToAddress(parts[0])
		amount := big.NewInt(0)
		amount.SetString(parts[1], 10)
		ret[addr] = core.GenesisAccount{Balance: amount}
	}
	return ret
}
