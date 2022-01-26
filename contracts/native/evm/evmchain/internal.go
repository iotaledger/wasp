// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmchain

import (
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/iotaledger/wasp/contracts/native/evm"
	"github.com/iotaledger/wasp/contracts/native/evm/evmchain/emulator"
	"github.com/iotaledger/wasp/contracts/native/evm/evminternal"
	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/assert"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

func isNotFound(err error) bool {
	switch err {
	case ethereum.NotFound:
		return true
	case emulator.ErrTransactionDoesNotExist:
		return true
	case emulator.ErrBlockDoesNotExist:
		return true
	}
	return false
}

// getEmulatorInBlockContext creates a new emulator instance if this is the first call to applyTransaction
// in the ISCP block; otherwise it returns the previously created instance. The purpose is to
// create a single Ethereum block for each ISCP block.
func getEmulatorInBlockContext(ctx iscp.Sandbox) *emulator.EVMEmulator {
	bctx := ctx.Privileged().BlockContext(createEmulator, commitEthereumBlock)
	return bctx.(*emulator.EVMEmulator)
}

func createEmulator(ctx iscp.Sandbox) interface{} {
	return emulator.NewEVMEmulator(rawdb.NewDatabase(emulator.NewKVAdapter(evminternal.EVMStateSubrealm(ctx.State()))), timestamp(ctx))
}

// timestamp returns the current timestamp in seconds since epoch
func timestamp(ctx iscp.SandboxBase) uint64 {
	tsNano := time.Duration(ctx.Timestamp()) * time.Nanosecond
	return uint64(tsNano / time.Second)
}

func commitEthereumBlock(blockContext interface{}) {
	emu := blockContext.(*emulator.EVMEmulator)
	emu.Commit()
	_ = emu.Close()
}

func withEmulatorR(ctx iscp.SandboxView, f func(*emulator.EVMEmulator) (dict.Dict, error)) (dict.Dict, error) {
	emu := emulator.NewEVMEmulator(
		rawdb.NewDatabase(emulator.NewKVAdapter(evminternal.EVMStateSubrealm(buffered.NewBufferedKVStoreAccess(ctx.State())))),
		timestamp(ctx),
	)
	defer emu.Close()
	return f(emu)
}

func txResult(ctx iscp.SandboxView, emu *emulator.EVMEmulator, tx *types.Transaction) dict.Dict {
	a := assert.NewAssert(ctx.Log())
	if tx == nil {
		return nil
	}
	receipt, err := emu.TransactionReceipt(tx.Hash())
	a.RequireNoError(err)
	return dict.Dict{
		evm.FieldTransaction:      evmtypes.EncodeTransaction(tx),
		evm.FieldBlockHash:        receipt.BlockHash.Bytes(),
		evm.FieldBlockNumber:      codec.EncodeUint64(receipt.BlockNumber.Uint64()),
		evm.FieldTransactionIndex: codec.EncodeUint64(uint64(receipt.TransactionIndex)),
	}
}

func withBlockByNumber(ctx iscp.SandboxView, f func(*emulator.EVMEmulator, *types.Block) (dict.Dict, error)) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	blockNumber := paramBlockNumber(ctx)

	return withEmulatorR(ctx, func(emu *emulator.EVMEmulator) (dict.Dict, error) {
		block, err := emu.BlockByNumber(blockNumber)
		if !isNotFound(err) {
			a.RequireNoError(err)
		}

		return f(emu, block)
	})
}

func withBlockByHash(ctx iscp.SandboxView, f func(*emulator.EVMEmulator, *types.Block) (dict.Dict, error)) (dict.Dict, error) {
	hash := common.BytesToHash(ctx.Params().MustGet(evm.FieldBlockHash))

	return withEmulatorR(ctx, func(emu *emulator.EVMEmulator) (dict.Dict, error) {
		block := emu.BlockByHash(hash)
		return f(emu, block)
	})
}

func withTransactionByHash(ctx iscp.SandboxView, f func(*emulator.EVMEmulator, *types.Transaction) (dict.Dict, error)) (dict.Dict, error) {
	txHash := common.BytesToHash(ctx.Params().MustGet(evm.FieldTransactionHash))

	return withEmulatorR(ctx, func(emu *emulator.EVMEmulator) (dict.Dict, error) {
		tx := emu.TransactionByHash(txHash)
		return f(emu, tx)
	})
}

func withTransactionByBlockHashAndIndex(ctx iscp.SandboxView, f func(*emulator.EVMEmulator, *types.Transaction) (dict.Dict, error)) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	blockHash := common.BytesToHash(ctx.Params().MustGet(evm.FieldBlockHash))
	index, err := codec.DecodeUint64(ctx.Params().MustGet(evm.FieldTransactionIndex), 0)
	a.RequireNoError(err)

	return withEmulatorR(ctx, func(emu *emulator.EVMEmulator) (dict.Dict, error) {
		tx, err := emu.TransactionInBlock(blockHash, uint(index))
		if !isNotFound(err) {
			a.RequireNoError(err)
		}
		return f(emu, tx)
	})
}

func withTransactionByBlockNumberAndIndex(ctx iscp.SandboxView, f func(*emulator.EVMEmulator, *types.Transaction) (dict.Dict, error)) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	index, err := codec.DecodeUint64(ctx.Params().MustGet(evm.FieldTransactionIndex), 0)
	a.RequireNoError(err)
	return withBlockByNumber(ctx, func(emu *emulator.EVMEmulator, block *types.Block) (dict.Dict, error) {
		if block == nil || index >= uint64(len(block.Transactions())) {
			return f(emu, nil)
		}
		return f(emu, block.Transactions()[index])
	})
}

func paramBlockNumber(ctx iscp.SandboxView) *big.Int {
	if ctx.Params().MustHas(evm.FieldBlockNumber) {
		return new(big.Int).SetBytes(ctx.Params().MustGet(evm.FieldBlockNumber))
	}
	return nil // latest block
}

func paramBlockNumberOrHashAsNumber(ctx iscp.SandboxView, emu *emulator.EVMEmulator) *big.Int {
	if ctx.Params().MustHas(evm.FieldBlockHash) {
		a := assert.NewAssert(ctx.Log())
		blockHash := common.BytesToHash(ctx.Params().MustGet(evm.FieldBlockHash))
		header, err := emu.HeaderByHash(blockHash)
		a.RequireNoError(err)
		a.Requiref(header != nil, "block not found")
		return header.Number
	}
	return paramBlockNumber(ctx)
}
