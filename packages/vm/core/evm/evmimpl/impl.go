// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/emulator"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

var Processor = evm.Contract.Processor(initialize,
	evm.FuncSetGasRatio.WithHandler(setGasRatio),
	evm.FuncGetGasRatio.WithHandler(getGasRatio),
	evm.FuncSetBlockTime.WithHandler(setBlockTime),
	evm.FuncMintBlock.WithHandler(mintBlock),
	evm.FuncSendTransaction.WithHandler(applyTransaction),
	evm.FuncGetBalance.WithHandler(getBalance),
	evm.FuncCallContract.WithHandler(callContract),
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
)

func initialize(ctx iscp.Sandbox) dict.Dict {
	genesisAlloc := core.GenesisAlloc{}
	var err error
	if ctx.Params().MustHas(evm.FieldGenesisAlloc) {
		genesisAlloc, err = evmtypes.DecodeGenesisAlloc(ctx.Params().MustGet(evm.FieldGenesisAlloc))
		ctx.RequireNoError(err)
	}

	gasLimit, err := codec.DecodeUint64(ctx.Params().MustGet(evm.FieldBlockGasLimit), evm.BlockGasLimitDefault)
	ctx.RequireNoError(err)

	blockKeepAmount, err := codec.DecodeInt32(ctx.Params().MustGet(evm.FieldBlockKeepAmount), evm.BlockKeepAmountDefault)
	ctx.RequireNoError(err)

	// add the standard ISC contract at arbitrary address 0x1074
	deployISCContractOnGenesis(genesisAlloc)

	chainID, err := codec.DecodeUint16(ctx.Params().MustGet(evm.FieldChainID), evm.DefaultChainID)
	ctx.RequireNoError(err)
	emulator.Init(
		evmStateSubrealm(ctx.State()),
		chainID,
		blockKeepAmount,
		gasLimit,
		timestamp(ctx),
		genesisAlloc,
	)

	gasRatio := codec.MustDecodeRatio32(ctx.Params().MustGet(evm.FieldGasRatio), evm.DefaultGasRatio)
	ctx.State().Set(keyGasRatio, gasRatio.Bytes())
	// storing hname as a terminal value of the contract's state nil key.
	// This way we will be able to retrieve commitment to the contract's state
	ctx.State().Set("", ctx.Contract().Bytes())

	return nil
}

func mintBlock(ctx iscp.Sandbox) dict.Dict {
	scheduleNextBlock(ctx)
	emu := createEmulator(ctx)
	emu.MintBlock()
	return nil
}

func applyTransaction(ctx iscp.Sandbox) dict.Dict {
	tx := &types.Transaction{}
	err := tx.UnmarshalBinary(ctx.Params().MustGet(evm.FieldTransactionData))
	ctx.RequireNoError(err)

	blockTime := getBlockTime(ctx.State())

	gasRatio := codec.MustDecodeRatio32(ctx.State().MustGet(keyGasRatio), evm.DefaultGasRatio)
	gasBudget := evm.ISCGasBudgetToEVM(ctx.Gas().Budget(), gasRatio)

	var emu *emulator.EVMEmulator
	if blockTime > 0 {
		// next block will be minted when mintBlock() is called (via timelocked request)
		emu = createEmulator(ctx)
	} else {
		// next block will be minted when the ISC block is closed
		emu = getEmulatorInBlockContext(ctx)
	}

	receipt, gasUsed, err, result := emu.SendTransaction(tx, gasBudget)

	// burn gas even on error
	ctx.Gas().Burn(gas.BurnCodeEVM1P, evm.EVMGasToISC(gasUsed, gasRatio))

	ctx.RequireNoError(err)

	// if EVM execution was reverted we must revert the ISC request as well
	ctx.Requiref(receipt.Status == types.ReceiptStatusSuccessful, GetRevertErrorMessage(result, ctx.Contract()))

	return nil
}

func getBalance(ctx iscp.SandboxView) dict.Dict {
	addr := common.BytesToAddress(ctx.Params().MustGet(evm.FieldAddress))
	emu := createEmulatorR(ctx)
	_ = paramBlockNumberOrHashAsNumber(ctx, emu, false)
	return result(emu.StateDB().GetBalance(addr).Bytes())
}

func getBlockNumber(ctx iscp.SandboxView) dict.Dict {
	emu := createEmulatorR(ctx)
	return result(new(big.Int).SetUint64(emu.BlockchainDB().GetNumber()).Bytes())
}

func getBlockByNumber(ctx iscp.SandboxView) dict.Dict {
	return blockResult(blockByNumber(ctx))
}

func getBlockByHash(ctx iscp.SandboxView) dict.Dict {
	return blockResult(blockByHash(ctx))
}

func getTransactionByHash(ctx iscp.SandboxView) dict.Dict {
	return txResult(transactionByHash(ctx))
}

func getTransactionByBlockHashAndIndex(ctx iscp.SandboxView) dict.Dict {
	return txResult(transactionByBlockHashAndIndex(ctx))
}

func getTransactionByBlockNumberAndIndex(ctx iscp.SandboxView) dict.Dict {
	return txResult(transactionByBlockNumberAndIndex(ctx))
}

func getTransactionCountByBlockHash(ctx iscp.SandboxView) dict.Dict {
	return txCountResult(blockByHash(ctx))
}

func getTransactionCountByBlockNumber(ctx iscp.SandboxView) dict.Dict {
	return txCountResult(blockByNumber(ctx))
}

func getReceipt(ctx iscp.SandboxView) dict.Dict {
	txHash := common.BytesToHash(ctx.Params().MustGet(evm.FieldTransactionHash))
	emu := createEmulatorR(ctx)
	r := emu.BlockchainDB().GetReceiptByTxHash(txHash)
	if r == nil {
		return nil
	}
	return result(evmtypes.EncodeReceiptFull(r))
}

func getNonce(ctx iscp.SandboxView) dict.Dict {
	emu := createEmulatorR(ctx)
	addr := common.BytesToAddress(ctx.Params().MustGet(evm.FieldAddress))
	_ = paramBlockNumberOrHashAsNumber(ctx, emu, false)
	return result(codec.EncodeUint64(emu.StateDB().GetNonce(addr)))
}

func getCode(ctx iscp.SandboxView) dict.Dict {
	emu := createEmulatorR(ctx)
	addr := common.BytesToAddress(ctx.Params().MustGet(evm.FieldAddress))
	_ = paramBlockNumberOrHashAsNumber(ctx, emu, false)
	return result(emu.StateDB().GetCode(addr))
}

func getStorage(ctx iscp.SandboxView) dict.Dict {
	emu := createEmulatorR(ctx)
	addr := common.BytesToAddress(ctx.Params().MustGet(evm.FieldAddress))
	key := common.BytesToHash(ctx.Params().MustGet(evm.FieldKey))
	_ = paramBlockNumberOrHashAsNumber(ctx, emu, false)
	data := emu.StateDB().GetState(addr, key)
	return result(data[:])
}

func getLogs(ctx iscp.SandboxView) dict.Dict {
	q, err := evmtypes.DecodeFilterQuery(ctx.Params().MustGet(evm.FieldFilterQuery))
	ctx.RequireNoError(err)
	emu := createEmulatorR(ctx)
	logs := emu.FilterLogs(q)
	return result(evmtypes.EncodeLogs(logs))
}

func callContract(ctx iscp.SandboxView) dict.Dict {
	callMsg, err := evmtypes.DecodeCallMsg(ctx.Params().MustGet(evm.FieldCallMsg))
	ctx.RequireNoError(err)
	emu := createEmulatorR(ctx)
	_ = paramBlockNumberOrHashAsNumber(ctx, emu, false)
	res, err := emu.CallContract(callMsg)
	ctx.RequireNoError(err)
	return result(res)
}
