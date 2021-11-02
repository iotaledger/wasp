// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmlight

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/iotaledger/wasp/contracts/native/evm"
	"github.com/iotaledger/wasp/contracts/native/evm/evminternal"
	"github.com/iotaledger/wasp/contracts/native/evm/evmlight/emulator"
	"github.com/iotaledger/wasp/contracts/native/evm/evmlight/iscpcontract"
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

func initialize(ctx iscp.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	genesisAlloc, err := evmtypes.DecodeGenesisAlloc(ctx.Params().MustGet(evm.FieldGenesisAlloc))
	a.RequireNoError(err)

	// add the standard ISCP contract at arbitrary address 0x1074
	iscpcontract.DeployOnGenesis(genesisAlloc, ctx.ChainID())

	chainID, err := codec.DecodeUint16(ctx.Params().MustGet(evm.FieldChainID), evm.DefaultChainID)
	a.RequireNoError(err)
	emulator.Init(
		evmStateSubrealm(ctx.State()),
		chainID,
		evm.GasLimitDefault, // TODO: make gas limit configurable
		timestamp(ctx),
		genesisAlloc,
	)
	evminternal.InitializeManagement(ctx)
	return nil, nil
}

func applyTransaction(ctx iscp.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())

	tx := &types.Transaction{}
	err := tx.UnmarshalBinary(ctx.Params().MustGet(evm.FieldTransactionData))
	a.RequireNoError(err)

	return evminternal.RequireGasFee(ctx, tx.Gas(), func() uint64 {
		emu := createEmulator(ctx)
		receipt, err := emu.SendTransaction(tx)
		a.RequireNoError(err)
		return receipt.GasUsed
	}), nil
}

func getBalance(ctx iscp.SandboxView) (dict.Dict, error) {
	addr := common.BytesToAddress(ctx.Params().MustGet(evm.FieldAddress))
	emu := createEmulatorR(ctx)
	_ = paramBlockNumberOrHashAsNumber(ctx, emu, false)
	return evminternal.Result(emu.StateDB().GetBalance(addr).Bytes()), nil
}

func getBlockNumber(ctx iscp.SandboxView) (dict.Dict, error) {
	emu := createEmulatorR(ctx)
	return evminternal.Result(emu.BlockchainDB().GetNumber().Bytes()), nil
}

func getBlockByNumber(ctx iscp.SandboxView) (dict.Dict, error) {
	return blockResult(blockByNumber(ctx))
}

func getBlockByHash(ctx iscp.SandboxView) (dict.Dict, error) {
	return blockResult(blockByHash(ctx))
}

func getTransactionByHash(ctx iscp.SandboxView) (dict.Dict, error) {
	return txResult(transactionByHash(ctx))
}

func getTransactionByBlockHashAndIndex(ctx iscp.SandboxView) (dict.Dict, error) {
	return txResult(transactionByBlockHashAndIndex(ctx))
}

func getTransactionByBlockNumberAndIndex(ctx iscp.SandboxView) (dict.Dict, error) {
	return txResult(transactionByBlockNumberAndIndex(ctx))
}

func getTransactionCountByBlockHash(ctx iscp.SandboxView) (dict.Dict, error) {
	return txCountResult(blockByHash(ctx))
}

func getTransactionCountByBlockNumber(ctx iscp.SandboxView) (dict.Dict, error) {
	return txCountResult(blockByNumber(ctx))
}

func getReceipt(ctx iscp.SandboxView) (dict.Dict, error) {
	txHash := common.BytesToHash(ctx.Params().MustGet(evm.FieldTransactionHash))
	emu := createEmulatorR(ctx)
	r := emu.BlockchainDB().GetReceiptByTxHash(txHash)
	if r == nil {
		return nil, nil
	}
	return evminternal.Result(evmtypes.EncodeReceiptFull(r)), nil
}

func getNonce(ctx iscp.SandboxView) (dict.Dict, error) {
	emu := createEmulatorR(ctx)
	addr := common.BytesToAddress(ctx.Params().MustGet(evm.FieldAddress))
	_ = paramBlockNumberOrHashAsNumber(ctx, emu, false)
	return evminternal.Result(codec.EncodeUint64(emu.StateDB().GetNonce(addr))), nil
}

func getCode(ctx iscp.SandboxView) (dict.Dict, error) {
	emu := createEmulatorR(ctx)
	addr := common.BytesToAddress(ctx.Params().MustGet(evm.FieldAddress))
	_ = paramBlockNumberOrHashAsNumber(ctx, emu, false)
	return evminternal.Result(emu.StateDB().GetCode(addr)), nil
}

func getStorage(ctx iscp.SandboxView) (dict.Dict, error) {
	emu := createEmulatorR(ctx)
	addr := common.BytesToAddress(ctx.Params().MustGet(evm.FieldAddress))
	key := common.BytesToHash(ctx.Params().MustGet(evm.FieldKey))
	_ = paramBlockNumberOrHashAsNumber(ctx, emu, false)
	data := emu.StateDB().GetState(addr, key)
	return evminternal.Result(data[:]), nil
}

func getLogs(ctx iscp.SandboxView) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	q, err := evmtypes.DecodeFilterQuery(ctx.Params().MustGet(evm.FieldFilterQuery))
	a.RequireNoError(err)
	emu := createEmulatorR(ctx)
	logs := emu.FilterLogs(q)
	return evminternal.Result(evmtypes.EncodeLogs(logs)), nil
}

func callContract(ctx iscp.SandboxView) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	callMsg, err := evmtypes.DecodeCallMsg(ctx.Params().MustGet(evm.FieldCallMsg))
	a.RequireNoError(err)
	emu := createEmulatorR(ctx)
	_ = paramBlockNumberOrHashAsNumber(ctx, emu, false)
	res, err := emu.CallContract(callMsg)
	a.RequireNoError(err)
	return evminternal.Result(res), nil
}

func estimateGas(ctx iscp.SandboxView) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	callMsg, err := evmtypes.DecodeCallMsg(ctx.Params().MustGet(evm.FieldCallMsg))
	a.RequireNoError(err)
	emu := createEmulatorR(ctx)
	gas, err := emu.EstimateGas(callMsg)
	a.RequireNoError(err)
	return evminternal.Result(codec.EncodeUint64(gas)), nil
}
