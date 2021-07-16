// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmchain

import (
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/evm"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/assert"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

func isNotFound(err error) bool {
	switch err {
	case ethereum.NotFound:
		return true
	case evm.ErrTransactionDoesNotExist:
		return true
	case evm.ErrBlockDoesNotExist:
		return true
	}
	return false
}

// getOrCreateEmulator creates a new emulator instance if this is the first call to applyTransaction
// in the ISCP block; otherwise it returns the previously created instance. The purpose is to
// create a single Ethereum block for each ISCP block.
func getOrCreateEmulator(ctx iscp.Sandbox) *evm.EVMEmulator {
	bctx := ctx.BlockContext(createEmulator, commitEthereumBlock)
	return bctx.(*evm.EVMEmulator)
}

func createEmulator(ctx iscp.Sandbox) interface{} {
	return evm.NewEVMEmulator(rawdb.NewDatabase(evm.NewKVAdapter(ctx.State())))
}

func commitEthereumBlock(blockContext interface{}) {
	emu := blockContext.(*evm.EVMEmulator)
	if emu.HasPendingBlock() {
		emu.Commit()
	}
	emu.Close()
}

func withEmulatorR(ctx iscp.SandboxView, f func(*evm.EVMEmulator) dict.Dict) (dict.Dict, error) {
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

func txResult(ctx iscp.SandboxView, emu *evm.EVMEmulator, tx *types.Transaction) dict.Dict {
	a := assert.NewAssert(ctx.Log())
	if tx == nil {
		return nil
	}
	receipt, err := emu.TransactionReceipt(tx.Hash())
	a.RequireNoError(err)
	return dict.Dict{
		FieldTransaction:      EncodeTransaction(tx),
		FieldBlockHash:        receipt.BlockHash.Bytes(),
		FieldBlockNumber:      codec.EncodeUint64(receipt.BlockNumber.Uint64()),
		FieldTransactionIndex: codec.EncodeUint64(uint64(receipt.TransactionIndex)),
	}
}

func withBlockByNumber(ctx iscp.SandboxView, f func(*evm.EVMEmulator, *types.Block) dict.Dict) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	blockNumber := paramBlockNumber(ctx)

	return withEmulatorR(ctx, func(emu *evm.EVMEmulator) dict.Dict {
		block, err := emu.BlockByNumber(blockNumber)
		if !isNotFound(err) {
			a.RequireNoError(err)
		}

		return f(emu, block)
	})
}

func withBlockByHash(ctx iscp.SandboxView, f func(*evm.EVMEmulator, *types.Block) dict.Dict) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	hash := common.BytesToHash(ctx.Params().MustGet(FieldBlockHash))

	return withEmulatorR(ctx, func(emu *evm.EVMEmulator) dict.Dict {
		block, err := emu.BlockByHash(hash)
		if !errors.Is(err, evm.ErrBlockDoesNotExist) {
			a.RequireNoError(err)
		}
		return f(emu, block)
	})
}

func withTransactionByHash(ctx iscp.SandboxView, f func(*evm.EVMEmulator, *types.Transaction) dict.Dict) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	txHash := common.BytesToHash(ctx.Params().MustGet(FieldTransactionHash))

	return withEmulatorR(ctx, func(emu *evm.EVMEmulator) dict.Dict {
		tx, pending, err := emu.TransactionByHash(txHash)
		a.Require(!pending, "unexpected pending transaction")
		if !isNotFound(err) {
			a.RequireNoError(err)
		}
		return f(emu, tx)
	})
}

func withTransactionByBlockHashAndIndex(ctx iscp.SandboxView, f func(*evm.EVMEmulator, *types.Transaction) dict.Dict) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	blockHash := common.BytesToHash(ctx.Params().MustGet(FieldBlockHash))
	index, _, err := codec.DecodeUint64(ctx.Params().MustGet(FieldTransactionIndex))
	a.RequireNoError(err)

	return withEmulatorR(ctx, func(emu *evm.EVMEmulator) dict.Dict {
		tx, err := emu.TransactionInBlock(blockHash, uint(index))
		if !isNotFound(err) {
			a.RequireNoError(err)
		}
		return f(emu, tx)
	})
}

func withTransactionByBlockNumberAndIndex(ctx iscp.SandboxView, f func(*evm.EVMEmulator, *types.Transaction) dict.Dict) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	index, _, err := codec.DecodeUint64(ctx.Params().MustGet(FieldTransactionIndex))
	a.RequireNoError(err)
	return withBlockByNumber(ctx, func(emu *evm.EVMEmulator, block *types.Block) dict.Dict {
		if block == nil || index >= uint64(len(block.Transactions())) {
			return f(emu, nil)
		}
		return f(emu, block.Transactions()[index])
	})
}

func paramBlockNumber(ctx iscp.SandboxView) *big.Int {
	if ctx.Params().MustHas(FieldBlockNumber) {
		return new(big.Int).SetBytes(ctx.Params().MustGet(FieldBlockNumber))
	}
	return nil
}

func getFeeColor(ctx iscp.Sandbox) ledgerstate.Color {
	a := assert.NewAssert(ctx.Log())

	// call root contract view to get the feecolor
	feeInfo, err := ctx.Call(
		root.Contract.Hname(),
		root.FuncGetFeeInfo.Hname(),
		dict.Dict{root.ParamHname: Contract.Hname().Bytes()},
		nil,
	)
	a.RequireNoError(err)
	feeColor, _, err := codec.DecodeColor(feeInfo.MustGet(root.ParamFeeColor))
	a.RequireNoError(err)
	return feeColor
}
