// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmchain

import (
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/assert"
	"github.com/iotaledger/wasp/packages/evm"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

func emulator(state kv.KVStore) *evm.EVMEmulator {
	return evm.NewEVMEmulator(rawdb.NewDatabase(evm.NewKVAdapter(state)))
}

func withEmulatorR(ctx coretypes.SandboxView, f func(*evm.EVMEmulator) dict.Dict) (dict.Dict, error) {
	emu := evm.NewEVMEmulator(rawdb.NewDatabase(evm.NewKVAdapter(buffered.NewBufferedKVStore(ctx.State()))))
	defer emu.Close()
	return f(emu), nil
}

func result(value []byte) dict.Dict {
	if value == nil {
		return nil
	}
	return dict.Dict{FieldResult: value}
}

func withBlockByNumber(ctx coretypes.SandboxView, f func(*evm.EVMEmulator, *types.Block) dict.Dict) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	blockNumber := paramBlockNumber(ctx)

	return withEmulatorR(ctx, func(emu *evm.EVMEmulator) dict.Dict {
		block, err := emu.BlockByNumber(blockNumber)
		if err != evm.ErrBlockDoesNotExist {
			a.RequireNoError(err)
		}

		return f(emu, block)
	})
}

func withBlockByHash(ctx coretypes.SandboxView, f func(*evm.EVMEmulator, *types.Block) dict.Dict) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	hash := common.BytesToHash(ctx.Params().MustGet(FieldBlockHash))

	return withEmulatorR(ctx, func(emu *evm.EVMEmulator) dict.Dict {
		block, err := emu.BlockByHash(hash)
		if err != evm.ErrBlockDoesNotExist {
			a.RequireNoError(err)
		}
		return f(emu, block)
	})
}

func withTransactionByHash(ctx coretypes.SandboxView, f func(*evm.EVMEmulator, *types.Transaction) dict.Dict) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	txHash := common.BytesToHash(ctx.Params().MustGet(FieldTransactionHash))

	return withEmulatorR(ctx, func(emu *evm.EVMEmulator) dict.Dict {
		tx, pending, err := emu.TransactionByHash(txHash)
		a.Require(!pending, "unexpected pending transaction")
		if err != ethereum.NotFound {
			a.RequireNoError(err)
		}
		return f(emu, tx)
	})
}

func paramBlockNumber(ctx coretypes.SandboxView) *big.Int {
	if ctx.Params().MustHas(FieldBlockNumber) {
		return new(big.Int).SetBytes(ctx.Params().MustGet(FieldBlockNumber))
	}
	return nil
}
