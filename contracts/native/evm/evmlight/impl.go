// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmlight

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/iotaledger/wasp/contracts/native/evm/evmlight/iscpcontract"
	"github.com/iotaledger/wasp/contracts/native/evm/evmlight/evm"
	"github.com/iotaledger/wasp/packages/evm/evmtypes"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/assert"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

const (
	keyEVMState   = "s"
	keyGasPerIota = "g"
	keyEVMOwner   = "o"
)

var Processor = Contract.Processor(initialize,
	// Ethereum blockchain
	FuncSendTransaction.WithHandler(applyTransaction),
	FuncGetBalance.WithHandler(getBalance),
	FuncCallContract.WithHandler(callContract),
	FuncEstimateGas.WithHandler(estimateGas),
	FuncGetNonce.WithHandler(getNonce),
	FuncGetReceipt.WithHandler(getReceipt),
	FuncGetCode.WithHandler(getCode),
	FuncGetBlockNumber.WithHandler(getBlockNumber),
	FuncGetBlockByNumber.WithHandler(getBlockByNumber),
	FuncGetBlockByHash.WithHandler(getBlockByHash),
	FuncGetTransactionByHash.WithHandler(getTransactionByHash),
	FuncGetTransactionByBlockHash.WithHandler(getTransactionByBlockHash),
	FuncGetTransactionByBlockNumber.WithHandler(getTransactionByBlockNumber),
	FuncGetTransactionCountByBlockHash.WithHandler(getTransactionCountByBlockHash),
	FuncGetTransactionCountByBlockNumber.WithHandler(getTransactionCountByBlockNumber),
	FuncGetStorage.WithHandler(getStorage),
	FuncGetLogs.WithHandler(getLogs),

	// EVMchain SC management
	FuncSetNextOwner.WithHandler(setNextOwner),
	FuncClaimOwnership.WithHandler(claimOwnership),
	FuncSetGasPerIota.WithHandler(setGasPerIota),
	FuncWithdrawGasFees.WithHandler(withdrawGasFees),
	FuncGetOwner.WithHandler(getOwner),
	FuncGetGasPerIota.WithHandler(getGasPerIota),
)

func initialize(ctx iscp.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	genesisAlloc, err := evmtypes.DecodeGenesisAlloc(ctx.Params().MustGet(FieldGenesisAlloc))
	a.RequireNoError(err)

	// add the standard ISCP contract at arbitrary address 0x1074
	iscpcontract.DeployOnGenesis(genesisAlloc, ctx.ChainID())

	chainID, err := codec.DecodeUint16(ctx.Params().MustGet(FieldChainID), evm.DefaultChainID)
	a.RequireNoError(err)
	evm.Init(
		evmStateSubrealm(ctx.State()),
		chainID,
		evm.GasLimitDefault, // TODO: make gas limit configurable
		timestamp(ctx),
		genesisAlloc,
	)
	ctx.State().Set(keyGasPerIota, codec.EncodeUint64(DefaultGasPerIota))
	ctx.State().Set(keyEVMOwner, codec.EncodeAgentID(ctx.ContractCreator()))
	return nil, nil
}

func applyTransaction(ctx iscp.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())

	tx := &types.Transaction{}
	err := tx.UnmarshalBinary(ctx.Params().MustGet(FieldTransactionData))
	a.RequireNoError(err)

	transferredIotas := ctx.IncomingTransfer().Get(getFeeColor(ctx))
	gasPerIota, err := codec.DecodeUint64(ctx.State().MustGet(keyGasPerIota), 0)
	a.RequireNoError(err)

	a.Require(
		transferredIotas >= tx.Gas()/gasPerIota,
		"transferred tokens (%d) not enough to cover the gas limit set in the transaction (%d at %d gas per iota token)", transferredIotas, tx.Gas(), gasPerIota,
	)

	emu := createEmulator(ctx)
	receipt, err := emu.SendTransaction(tx)
	a.RequireNoError(err)

	iotasGasFee := receipt.GasUsed / gasPerIota
	if transferredIotas > iotasGasFee {
		// refund unspent gas fee to the sender's on-chain account
		iotasGasRefund := transferredIotas - iotasGasFee
		_, err = ctx.Call(
			accounts.Contract.Hname(),
			accounts.FuncDeposit.Hname(),
			dict.Dict{accounts.ParamAgentID: codec.EncodeAgentID(ctx.Caller())},
			colored.NewBalancesForIotas(iotasGasRefund),
		)
		a.RequireNoError(err)
	}

	return dict.Dict{
		FieldGasFee:  codec.EncodeUint64(iotasGasFee),
		FieldGasUsed: codec.EncodeUint64(receipt.GasUsed),
	}, nil
}

func getBalance(ctx iscp.SandboxView) (dict.Dict, error) {
	addr := common.BytesToAddress(ctx.Params().MustGet(FieldAddress))
	emu := createEmulatorR(ctx)
	_ = paramBlockNumber(ctx, emu, false)
	return result(emu.StateDB().GetBalance(addr).Bytes()), nil
}

func getBlockNumber(ctx iscp.SandboxView) (dict.Dict, error) {
	emu := createEmulatorR(ctx)
	return result(emu.BlockchainDB().GetNumber().Bytes()), nil
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

func getTransactionByBlockHash(ctx iscp.SandboxView) (dict.Dict, error) {
	return txResult(transactionByBlockHash(ctx))
}

func getTransactionByBlockNumber(ctx iscp.SandboxView) (dict.Dict, error) {
	return txResult(transactionByBlockNumber(ctx))
}

func getTransactionCountByBlockHash(ctx iscp.SandboxView) (dict.Dict, error) {
	return txCountResult(blockByHash(ctx))
}

func getTransactionCountByBlockNumber(ctx iscp.SandboxView) (dict.Dict, error) {
	return txCountResult(blockByNumber(ctx))
}

func getReceipt(ctx iscp.SandboxView) (dict.Dict, error) {
	txHash := common.BytesToHash(ctx.Params().MustGet(FieldTransactionHash))
	emu := createEmulatorR(ctx)
	r := emu.BlockchainDB().GetReceiptByTxHash(txHash)
	if r == nil {
		return nil, nil
	}
	return result(evmtypes.EncodeReceiptFull(r)), nil
}

func getNonce(ctx iscp.SandboxView) (dict.Dict, error) {
	emu := createEmulatorR(ctx)
	addr := common.BytesToAddress(ctx.Params().MustGet(FieldAddress))
	_ = paramBlockNumber(ctx, emu, false)
	return result(codec.EncodeUint64(emu.StateDB().GetNonce(addr))), nil
}

func getCode(ctx iscp.SandboxView) (dict.Dict, error) {
	emu := createEmulatorR(ctx)
	addr := common.BytesToAddress(ctx.Params().MustGet(FieldAddress))
	_ = paramBlockNumber(ctx, emu, false)
	return result(emu.StateDB().GetCode(addr)), nil
}

func getStorage(ctx iscp.SandboxView) (dict.Dict, error) {
	emu := createEmulatorR(ctx)
	addr := common.BytesToAddress(ctx.Params().MustGet(FieldAddress))
	key := common.BytesToHash(ctx.Params().MustGet(FieldKey))
	_ = paramBlockNumber(ctx, emu, false)
	data := emu.StateDB().GetState(addr, key)
	return result(data[:]), nil
}

func getLogs(ctx iscp.SandboxView) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	q, err := evmtypes.DecodeFilterQuery(ctx.Params().MustGet(FieldFilterQuery))
	a.RequireNoError(err)
	emu := createEmulatorR(ctx)
	logs := emu.FilterLogs(q)
	return result(evmtypes.EncodeLogs(logs)), nil
}

func callContract(ctx iscp.SandboxView) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	callMsg, err := evmtypes.DecodeCallMsg(ctx.Params().MustGet(FieldCallMsg))
	a.RequireNoError(err)
	emu := createEmulatorR(ctx)
	_ = paramBlockNumber(ctx, emu, false)
	res, err := emu.CallContract(callMsg)
	a.RequireNoError(err)
	return result(res), nil
}

func estimateGas(ctx iscp.SandboxView) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	callMsg, err := evmtypes.DecodeCallMsg(ctx.Params().MustGet(FieldCallMsg))
	a.RequireNoError(err)
	emu := createEmulatorR(ctx)
	gas, err := emu.EstimateGas(callMsg)
	a.RequireNoError(err)
	return result(codec.EncodeUint64(gas)), nil
}

// EVM chain management functions ///////////////////////////////////////////////////////////////////////////////////////

func requireOwner(ctx iscp.Sandbox) {
	contractOwner, err := codec.DecodeAgentID(ctx.State().MustGet(keyEVMOwner))
	a := assert.NewAssert(ctx.Log())
	a.RequireNoError(err)
	a.Require(contractOwner.Equals(ctx.Caller()), "can only be called by the contract owner")
}

func setNextOwner(ctx iscp.Sandbox) (dict.Dict, error) {
	requireOwner(ctx)
	par := kvdecoder.New(ctx.Params(), ctx.Log())
	ctx.State().Set(FieldNextEvmOwner, codec.EncodeAgentID(par.MustGetAgentID(FieldNextEvmOwner)))
	return nil, nil
}

func claimOwnership(ctx iscp.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())

	nextOwner, err := codec.DecodeAgentID(ctx.State().MustGet(FieldNextEvmOwner))
	a.RequireNoError(err)
	a.Require(nextOwner.Equals(ctx.Caller()), "Can only be called by the contract owner")

	ctx.State().Set(keyEVMOwner, codec.EncodeAgentID(nextOwner))
	return nil, nil
}

func getOwner(ctx iscp.SandboxView) (dict.Dict, error) {
	return result(ctx.State().MustGet(keyEVMOwner)), nil
}

func setGasPerIota(ctx iscp.Sandbox) (dict.Dict, error) {
	requireOwner(ctx)
	par := kvdecoder.New(ctx.Params())
	gasPerIotaBin := codec.EncodeUint64(par.MustGetUint64(FieldGasPerIota))
	ctx.State().Set(keyGasPerIota, gasPerIotaBin)
	return nil, nil
}

func getGasPerIota(ctx iscp.SandboxView) (dict.Dict, error) {
	return result(ctx.State().MustGet(keyGasPerIota)), nil
}

func withdrawGasFees(ctx iscp.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	requireOwner(ctx)

	paramsDecoder := kvdecoder.New(ctx.Params(), ctx.Log())
	targetAgentID := paramsDecoder.MustGetAgentID(FieldAgentID, ctx.Caller())

	isOnChain := targetAgentID.Address().Equals(ctx.ChainID().AsAddress())

	if isOnChain {
		params := codec.MakeDict(map[string]interface{}{
			accounts.ParamAgentID: targetAgentID,
		})
		_, err := ctx.Call(accounts.Contract.Hname(), accounts.FuncDeposit.Hname(), params, ctx.Balances())
		a.RequireNoError(err)
		return nil, nil
	}

	a.Require(ctx.Send(targetAgentID.Address(), ctx.Balances(), &iscp.SendMetadata{
		TargetContract: targetAgentID.Hname(),
	}), "withdraw.inconsistency: failed sending tokens ")

	return nil, nil
}
