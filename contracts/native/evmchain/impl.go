// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmchain

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/assert"
	"github.com/iotaledger/wasp/packages/evm"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

func initialize(ctx coretypes.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())

	// faucet address with initial supply
	faucet, err := crypto.GenerateKey()
	a.RequireNoError(err)
	faucetAddress := crypto.PubkeyToAddress(faucet.PublicKey)
	faucetSupply := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(9))
	genesisAlloc := map[common.Address]core.GenesisAccount{
		faucetAddress: {Balance: faucetSupply},
	}

	// this commits the genesis block
	emu := evm.NewEVMEmulator(rawdb.NewDatabase(evm.NewKVAdapter(ctx.State())), genesisAlloc)
	defer emu.Close()
	return nil, nil
}
