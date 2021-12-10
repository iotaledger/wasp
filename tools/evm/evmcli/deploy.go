// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmcli

import (
	"encoding/base64"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/iotaledger/wasp/contracts/native/evm"
	"github.com/iotaledger/wasp/contracts/native/evm/evmchain"
	"github.com/iotaledger/wasp/packages/evm/evmflavors"
	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/spf13/cobra"
)

type DeployParams struct {
	evmFlavor   string
	ChainID     int
	name        string
	description string
	alloc       []string
	allocBase64 string
	GasPerIOTA  uint64
}

func (d *DeployParams) InitFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&d.evmFlavor, "evm-flavor", "", evmchain.Contract.Name, "EVM flavor. One of `evmchain`, `evmlight`")
	cmd.Flags().IntVarP(&d.ChainID, "chainid", "", evm.DefaultChainID, "ChainID")
	cmd.Flags().StringVarP(&d.name, "name", "", "", "Contract name. Default: same as --evm-flavor")
	cmd.Flags().StringVarP(&d.description, "description", "", "", "Contract description")
	cmd.Flags().StringSliceVarP(&d.alloc, "alloc", "", nil, "Genesis allocation (format: <address>:<wei>,<address>:<wei>,...)")
	cmd.Flags().StringVarP(&d.allocBase64, "alloc-bytes", "", "", "Genesis allocation (base64-encoded)")
	cmd.Flags().Uint64VarP(&d.GasPerIOTA, "gas-per-iota", "", evm.DefaultGasPerIota, "Gas per IOTA charged as fee")
}

func (d *DeployParams) Name() string {
	if d.name != "" {
		return d.name
	}
	return d.EVMFlavor().Name
}

func (d *DeployParams) Description() string {
	if d.description != "" {
		return d.description
	}
	return d.EVMFlavor().Description
}

func (d *DeployParams) EVMFlavor() *coreutil.ContractInfo {
	r, ok := evmflavors.All[d.evmFlavor]
	if !ok {
		log.Fatalf("unknown EVM flavor: %s", d.evmFlavor)
	}
	return r
}

func (d *DeployParams) GetGenesis(def core.GenesisAlloc) core.GenesisAlloc {
	if len(d.alloc) != 0 && d.allocBase64 != "" {
		log.Fatalf("--alloc and --alloc-bytes are mutually exclusive")
	}
	if len(d.alloc) == 0 && d.allocBase64 == "" {
		if len(def) == 0 {
			log.Fatalf("One of --alloc and --alloc-bytes is mandatory")
		}
		return def
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
