// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmchain

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/iotaledger/wasp/contracts/native/evm"
	"github.com/iotaledger/wasp/contracts/native/evm/evmchain/emulator"
	"github.com/iotaledger/wasp/contracts/native/evm/evminternal"
	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/assert"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

var Processor = Contract.Processor(initialize, append(
	evminternal.ManagementHandlers,

	evm.FuncSendTransaction.WithHandler(applyTransaction),
	evm.FuncGetBalance.WithHandler(getBalance),
	evm.FuncCallContract.WithHandler(callContract),
	evm.FuncEstimateGas.WithHandler(estimateGas),
	evm.FuncGetNonce.WithHandler(getNonce),
	evm.FuncGetReceipt.WithHandler(getReceipt),
	evm.FuncGetCode.WithHandler(getCode),
	evm.FuncGetBlockNumber.WithHandler(getBlockNumber),
	evm.FuncGetBlockByNumber.WithHandler(getBlockByNumber),
	evm.FuncGetBlockByHash.WithHandler(getBlockByHash),
	evm.FuncGetTransactionByHash.WithHandler(getTransactionByHash),
	evm.FuncGetTransactionByBlockHashAndIndex.WithHandler(getTransactionByBlockHashAndIndex),
	evm.FuncGetTransactionByBlockNumberAndIndex.WithHandler(getTransactionByBlockNumberAndIndex),
	evm.FuncGetTransactionCountByBlockHash.WithHandler(getTransactionCountByBlockHash),
	evm.FuncGetTransactionCountByBlockNumber.WithHandler(getTransactionCountByBlockNumber),
	evm.FuncGetStorage.WithHandler(getStorage),
	evm.FuncGetLogs.WithHandler(getLogs),
)...)

func initialize(ctx iscp.Sandbox) dict.Dict {
	a := assert.NewAssert(ctx.Log())
	genesisAlloc, err := evmtypes.DecodeGenesisAlloc(ctx.Params().MustGet(evm.FieldGenesisAlloc))
	a.RequireNoError(err)

	gasLimit, err := codec.DecodeUint64(ctx.Params().MustGet(evm.FieldBlockGasLimit), evm.BlockGasLimitDefault)
	a.RequireNoError(err)

	chainID, err := codec.DecodeUint16(ctx.Params().MustGet(evm.FieldChainID), evm.DefaultChainID)
	a.RequireNoError(err)
	emulator.InitGenesis(
		int(chainID),
		rawdb.NewDatabase(emulator.NewKVAdapter(evminternal.EVMStateSubrealm(ctx.State()))),
		genesisAlloc,
		gasLimit,
		timestamp(ctx),
	)
	evminternal.InitializeManagement(ctx)
	return nil
}

func applyTransaction(ctx iscp.Sandbox) dict.Dict {
	return evminternal.ApplyTransaction(ctx, func(tx *types.Transaction, _ uint32, gasBudget uint64) (uint64, error) {
		_, gasUsed, error := getEmulatorInBlockContext(ctx).SendTransaction(tx, gasBudget)
		return gasUsed, error
	})
}

func getBalance(ctx iscp.SandboxView) dict.Dict {
	a := assert.NewAssert(ctx.Log())
	addr := common.BytesToAddress(ctx.Params().MustGet(evm.FieldAddress))

	return withEmulatorR(ctx, func(emu *emulator.EVMEmulator) dict.Dict {
		blockNumber := paramBlockNumberOrHashAsNumber(ctx, emu)
		bal, err := emu.BalanceAt(addr, blockNumber)
		a.RequireNoError(err)
		return evminternal.Result(bal.Bytes())
	})
}

func getBlockNumber(ctx iscp.SandboxView) dict.Dict {
	return withEmulatorR(ctx, func(emu *emulator.EVMEmulator) dict.Dict {
		return evminternal.Result(emu.Blockchain().CurrentBlock().Number().Bytes())
	})
}

func getBlockByNumber(ctx iscp.SandboxView) dict.Dict {
	return withBlockByNumber(ctx, func(emu *emulator.EVMEmulator, block *types.Block) dict.Dict {
		if block == nil {
			return nil
		}
		return evminternal.Result(evmtypes.EncodeBlock(block))
	})
}

func getBlockByHash(ctx iscp.SandboxView) dict.Dict {
	return withBlockByHash(ctx, func(emu *emulator.EVMEmulator, block *types.Block) dict.Dict {
		if block == nil {
			return nil
		}
		return evminternal.Result(evmtypes.EncodeBlock(block))
	})
}

func getTransactionByHash(ctx iscp.SandboxView) dict.Dict {
	return withTransactionByHash(ctx, func(emu *emulator.EVMEmulator, tx *types.Transaction) dict.Dict {
		return txResult(ctx, emu, tx)
	})
}

func getTransactionByBlockHashAndIndex(ctx iscp.SandboxView) dict.Dict {
	return withTransactionByBlockHashAndIndex(ctx, func(emu *emulator.EVMEmulator, tx *types.Transaction) dict.Dict {
		return txResult(ctx, emu, tx)
	})
}

func getTransactionByBlockNumberAndIndex(ctx iscp.SandboxView) dict.Dict {
	return withTransactionByBlockNumberAndIndex(ctx, func(emu *emulator.EVMEmulator, tx *types.Transaction) dict.Dict {
		return txResult(ctx, emu, tx)
	})
}

func getTransactionCountByBlockHash(ctx iscp.SandboxView) dict.Dict {
	return withBlockByHash(ctx, func(emu *emulator.EVMEmulator, block *types.Block) dict.Dict {
		if block == nil {
			return nil
		}
		return evminternal.Result(codec.EncodeUint64(uint64(len(block.Transactions()))))
	})
}

func getTransactionCountByBlockNumber(ctx iscp.SandboxView) dict.Dict {
	return withBlockByNumber(ctx, func(emu *emulator.EVMEmulator, block *types.Block) dict.Dict {
		if block == nil {
			return nil
		}
		return evminternal.Result(codec.EncodeUint64(uint64(len(block.Transactions()))))
	})
}

func getReceipt(ctx iscp.SandboxView) dict.Dict {
	a := assert.NewAssert(ctx.Log())
	return withTransactionByHash(ctx, func(emu *emulator.EVMEmulator, tx *types.Transaction) dict.Dict {
		if tx == nil {
			return nil
		}
		receipt, err := emu.TransactionReceipt(tx.Hash())
		a.RequireNoError(err)

		return evminternal.Result(evmtypes.EncodeReceiptFull(receipt))
	})
}

func getNonce(ctx iscp.SandboxView) dict.Dict {
	a := assert.NewAssert(ctx.Log())
	addr := common.BytesToAddress(ctx.Params().MustGet(evm.FieldAddress))

	return withEmulatorR(ctx, func(emu *emulator.EVMEmulator) dict.Dict {
		blockNumber := paramBlockNumberOrHashAsNumber(ctx, emu)
		nonce, err := emu.NonceAt(addr, blockNumber)
		a.RequireNoError(err)
		return evminternal.Result(codec.EncodeUint64(nonce))
	})
}

func getCode(ctx iscp.SandboxView) dict.Dict {
	a := assert.NewAssert(ctx.Log())
	addr := common.BytesToAddress(ctx.Params().MustGet(evm.FieldAddress))

	return withEmulatorR(ctx, func(emu *emulator.EVMEmulator) dict.Dict {
		blockNumber := paramBlockNumberOrHashAsNumber(ctx, emu)
		code, err := emu.CodeAt(addr, blockNumber)
		a.RequireNoError(err)
		return evminternal.Result(code)
	})
}

func getStorage(ctx iscp.SandboxView) dict.Dict {
	a := assert.NewAssert(ctx.Log())
	addr := common.BytesToAddress(ctx.Params().MustGet(evm.FieldAddress))
	key := common.BytesToHash(ctx.Params().MustGet(evm.FieldKey))

	return withEmulatorR(ctx, func(emu *emulator.EVMEmulator) dict.Dict {
		blockNumber := paramBlockNumberOrHashAsNumber(ctx, emu)
		data, err := emu.StorageAt(addr, key, blockNumber)
		a.RequireNoError(err)
		return evminternal.Result(data)
	})
}

func getLogs(ctx iscp.SandboxView) dict.Dict {
	a := assert.NewAssert(ctx.Log())
	q, err := evmtypes.DecodeFilterQuery(ctx.Params().MustGet(evm.FieldFilterQuery))
	a.RequireNoError(err)

	return withEmulatorR(ctx, func(emu *emulator.EVMEmulator) dict.Dict {
		logs, err := emu.FilterLogs(q)
		a.RequireNoError(err)
		return evminternal.Result(evmtypes.EncodeLogs(logs))
	})
}

func callContract(ctx iscp.SandboxView) dict.Dict {
	a := assert.NewAssert(ctx.Log())
	callMsg, err := evmtypes.DecodeCallMsg(ctx.Params().MustGet(evm.FieldCallMsg))
	a.RequireNoError(err)

	return withEmulatorR(ctx, func(emu *emulator.EVMEmulator) dict.Dict {
		blockNumber := paramBlockNumberOrHashAsNumber(ctx, emu)
		res, err := emu.CallContract(callMsg, blockNumber)
		a.RequireNoError(err)
		return evminternal.Result(res)
	})
}

func estimateGas(ctx iscp.SandboxView) dict.Dict {
	a := assert.NewAssert(ctx.Log())
	callMsg, err := evmtypes.DecodeCallMsg(ctx.Params().MustGet(evm.FieldCallMsg))
	a.RequireNoError(err)

	return withEmulatorR(ctx, func(emu *emulator.EVMEmulator) dict.Dict {
		gas, err := emu.EstimateGas(callMsg)
		a.RequireNoError(err)
		return evminternal.Result(codec.EncodeUint64(gas))
	})
}
