// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmchain

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/assert"
	"github.com/iotaledger/wasp/packages/evm"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

func initialize(ctx coretypes.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	genesisAlloc, err := DecodeGenesisAlloc(ctx.Params().MustGet(FieldGenesisAlloc))
	a.RequireNoError(err)
	evm.InitGenesis(rawdb.NewDatabase(evm.NewKVAdapter(ctx.State())), genesisAlloc)
	ctx.State().Set(FieldGasPerIota, codec.EncodeUint64(DefaultGasPerIota))    // sets the default GasPerIota value
	ctx.State().Set(FieldEvmOwner, codec.EncodeAgentID(ctx.ContractCreator())) // sets the default contract owner
	return nil, nil
}

func applyTransaction(ctx coretypes.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())

	tx := &types.Transaction{}
	err := tx.UnmarshalBinary(ctx.Params().MustGet(FieldTransactionData))
	a.RequireNoError(err)

	// call root contract view to get the feecolor
	params := codec.MakeDict(map[string]interface{}{
		root.ParamHname: coretypes.Hn(Name),
	})
	feeInfo, err := ctx.Call(root.Interface.Hname(), coretypes.Hn(root.FuncGetFeeInfo), params, nil)
	a.RequireNoError(err)
	feeColor, _, err := codec.DecodeColor(feeInfo.MustGet(root.ParamFeeColor))
	a.RequireNoError(err)

	transferredIotas, _ := ctx.IncomingTransfer().Get(feeColor)
	gasPerIota, _, err := codec.DecodeUint64(ctx.State().MustGet(FieldGasPerIota))
	a.RequireNoError(err)

	a.Require(
		transferredIotas >= tx.Gas()/gasPerIota,
		"transferred tokens (%d) not enough to cover the gas limit set in the transaction (%d at %d gas per iota token)", transferredIotas, tx.Gas(), gasPerIota,
	)

	// TODO: use block context
	emu := emulator(ctx.State())
	defer emu.Close()

	err = emu.SendTransaction(tx)
	if err != nil {
		ctx.Log().Infof(err.Error())
	}
	a.RequireNoError(err)

	// TODO: commit the pending ethereum block when closing the ISCP block
	emu.Commit()

	receipt, err := emu.TransactionReceipt(tx.Hash())
	a.RequireNoError(err)

	var signer types.Signer = types.FrontierSigner{}
	if tx.Protected() {
		signer = types.NewEIP155Signer(tx.ChainId())
	}
	from, _ := types.Sender(signer, tx)

	receiptBytes := NewReceipt(receipt, from, tx.To()).Bytes()

	iotasGasFee := receipt.GasUsed / gasPerIota

	a.Require(
		transferredIotas >= iotasGasFee,
		"transferred tokens (%d) not enough to pay the consumed gas fees (%d)", transferredIotas, iotasGasFee,
	)

	// refund unspent gas fee to the sender's on-chain account
	iotasGasRefund := transferredIotas - iotasGasFee
	params = codec.MakeDict(map[string]interface{}{
		accounts.ParamAgentID: ctx.Caller(),
	})
	_, err = ctx.Call(accounts.Interface.Hname(), coretypes.Hn(accounts.FuncDeposit), params, coretypes.NewTransferIotas(iotasGasRefund))
	a.RequireNoError(err)

	ret := dict.New()
	ret.Set(FieldResult, receiptBytes)
	ret.Set(FieldGasFee, codec.EncodeUint64(iotasGasFee))

	return ret, nil
}

func getBalance(ctx coretypes.SandboxView) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	addr := common.BytesToAddress(ctx.Params().MustGet(FieldAddress))
	blockNumber := paramBlockNumber(ctx)

	return withEmulatorR(ctx, func(emu *evm.EVMEmulator) dict.Dict {
		bal, err := emu.BalanceAt(addr, blockNumber)
		a.RequireNoError(err)
		return result(bal.Bytes())
	})
}

func getBlockNumber(ctx coretypes.SandboxView) (dict.Dict, error) {
	return withEmulatorR(ctx, func(emu *evm.EVMEmulator) dict.Dict {
		return result(emu.Blockchain().CurrentBlock().Number().Bytes())
	})
}

func getBlockByNumber(ctx coretypes.SandboxView) (dict.Dict, error) {
	return withBlockByNumber(ctx, func(emu *evm.EVMEmulator, block *types.Block) dict.Dict {
		if block == nil {
			return nil
		}
		return result(EncodeBlock(block))
	})
}

func getBlockByHash(ctx coretypes.SandboxView) (dict.Dict, error) {
	return withBlockByHash(ctx, func(emu *evm.EVMEmulator, block *types.Block) dict.Dict {
		if block == nil {
			return nil
		}
		return result(EncodeBlock(block))
	})
}

func getBlockTransactionCountByHash(ctx coretypes.SandboxView) (dict.Dict, error) {
	return withBlockByHash(ctx, func(emu *evm.EVMEmulator, block *types.Block) dict.Dict {
		if block == nil {
			return nil
		}
		return result(codec.EncodeUint64(uint64(len(block.Transactions()))))
	})
}

func getUncleCountByBlockHash(ctx coretypes.SandboxView) (dict.Dict, error) {
	return withBlockByHash(ctx, func(emu *evm.EVMEmulator, block *types.Block) dict.Dict {
		if block == nil {
			return nil
		}
		return result(codec.EncodeUint64(uint64(len(block.Uncles()))))
	})
}

func getBlockTransactionCountByNumber(ctx coretypes.SandboxView) (dict.Dict, error) {
	return withBlockByNumber(ctx, func(emu *evm.EVMEmulator, block *types.Block) dict.Dict {
		if block == nil {
			return nil
		}
		return result(codec.EncodeUint64(uint64(len(block.Transactions()))))
	})
}

func getUncleCountByBlockNumber(ctx coretypes.SandboxView) (dict.Dict, error) {
	return withBlockByNumber(ctx, func(emu *evm.EVMEmulator, block *types.Block) dict.Dict {
		if block == nil {
			return nil
		}
		return result(codec.EncodeUint64(uint64(len(block.Uncles()))))
	})
}

func getReceipt(ctx coretypes.SandboxView) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	txHash := common.BytesToHash(ctx.Params().MustGet(FieldTransactionHash))

	return withEmulatorR(ctx, func(emu *evm.EVMEmulator) dict.Dict {
		receipt, err := emu.TransactionReceipt(txHash)
		a.RequireNoError(err)

		block := emu.Blockchain().GetBlockByHash(receipt.BlockHash)
		if block == nil {
			ctx.Log().Panicf("block for receipt not found")
		}

		if receipt.TransactionIndex >= uint(len(block.Transactions())) {
			ctx.Log().Panicf("cannot find transaction in block")
		}
		tx := block.Transactions()[receipt.TransactionIndex]

		var signer types.Signer = types.FrontierSigner{}
		if tx.Protected() {
			signer = types.NewEIP155Signer(tx.ChainId())
		}
		from, _ := types.Sender(signer, tx)

		return result(NewReceipt(receipt, from, tx.To()).Bytes())
	})
}

func getNonce(ctx coretypes.SandboxView) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	addr := common.BytesToAddress(ctx.Params().MustGet(FieldAddress))
	blockNumber := paramBlockNumber(ctx)

	return withEmulatorR(ctx, func(emu *evm.EVMEmulator) dict.Dict {
		nonce, err := emu.NonceAt(addr, blockNumber)
		a.RequireNoError(err)
		return result(codec.EncodeUint64(nonce))
	})
}

func getCode(ctx coretypes.SandboxView) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	addr := common.BytesToAddress(ctx.Params().MustGet(FieldAddress))
	blockNumber := paramBlockNumber(ctx)

	return withEmulatorR(ctx, func(emu *evm.EVMEmulator) dict.Dict {
		code, err := emu.CodeAt(addr, blockNumber)
		a.RequireNoError(err)
		return result(code)
	})
}

func getStorage(ctx coretypes.SandboxView) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	addr := common.BytesToAddress(ctx.Params().MustGet(FieldAddress))
	key := common.BytesToHash(ctx.Params().MustGet(FieldKey))
	blockNumber := paramBlockNumber(ctx)

	return withEmulatorR(ctx, func(emu *evm.EVMEmulator) dict.Dict {
		data, err := emu.StorageAt(addr, key, blockNumber)
		a.RequireNoError(err)
		return result(data)
	})
}

func callContract(ctx coretypes.SandboxView) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	callMsg, err := DecodeCallMsg(ctx.Params().MustGet(FieldCallMsg))
	a.RequireNoError(err)
	blockNumber := paramBlockNumber(ctx)

	return withEmulatorR(ctx, func(emu *evm.EVMEmulator) dict.Dict {
		res, err := emu.CallContract(callMsg, blockNumber)
		a.RequireNoError(err)
		return result(res)
	})
}

func estimateGas(ctx coretypes.SandboxView) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	callMsg, err := DecodeCallMsg(ctx.Params().MustGet(FieldCallMsg))
	a.RequireNoError(err)

	return withEmulatorR(ctx, func(emu *evm.EVMEmulator) dict.Dict {
		gas, err := emu.EstimateGas(callMsg)
		a.RequireNoError(err)
		return result(codec.EncodeUint64(gas))
	})
}

// EVM chain management functions ///////////////////////////////////////////////////////////////////////////////////////

func requireOwner(ctx coretypes.Sandbox) {
	contractOwner, _, err := codec.DecodeAgentID(ctx.State().MustGet(FieldEvmOwner))
	a := assert.NewAssert(ctx.Log())
	a.RequireNoError(err)
	a.Require(contractOwner.Equals(ctx.Caller()), "can only be called by the contract owner")
}

func setNextOwner(ctx coretypes.Sandbox) (dict.Dict, error) {
	requireOwner(ctx)
	par := kvdecoder.New(ctx.Params(), ctx.Log())
	ctx.State().Set(FieldNextEvmOwner, codec.EncodeAgentID(par.MustGetAgentID(FieldNextEvmOwner)))
	return nil, nil
}

func claimOwnership(ctx coretypes.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())

	nextOwner, _, err := codec.DecodeAgentID(ctx.State().MustGet(FieldNextEvmOwner))
	a.RequireNoError(err)
	a.Require(nextOwner.Equals(ctx.Caller()), "Can only be called by the contract owner")

	ctx.State().Set(FieldEvmOwner, codec.EncodeAgentID(&nextOwner))
	return nil, nil
}

func getOwner(ctx coretypes.SandboxView) (dict.Dict, error) {
	return result(ctx.State().MustGet(FieldEvmOwner)), nil
}

func setGasPerIota(ctx coretypes.Sandbox) (dict.Dict, error) {
	requireOwner(ctx)
	par := kvdecoder.New(ctx.Params())
	gasPerIotaBin := codec.EncodeUint64(par.MustGetUint64(FieldGasPerIota))
	ctx.State().Set(FieldGasPerIota, gasPerIotaBin)
	return nil, nil
}

func getGasPerIota(ctx coretypes.SandboxView) (dict.Dict, error) {
	return result(ctx.State().MustGet(FieldGasPerIota)), nil
}

func withdrawGasFees(ctx coretypes.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	requireOwner(ctx)

	paramsDecoder := kvdecoder.New(ctx.Params(), ctx.Log())
	targetAgentId := paramsDecoder.MustGetAgentID(FieldAgentId, *ctx.Caller())

	isOnChain := targetAgentId.Address().Equals(ctx.ChainID().AsAddress())

	if isOnChain {
		params := codec.MakeDict(map[string]interface{}{
			accounts.ParamAgentID: targetAgentId,
		})
		_, err := ctx.Call(accounts.Interface.Hname(), coretypes.Hn(accounts.FuncDeposit), params, ctx.Balances())
		a.RequireNoError(err)
		return nil, nil
	}

	a.Require(ctx.Send(targetAgentId.Address(), ctx.Balances(), &coretypes.SendMetadata{
		TargetContract: targetAgentId.Hname(),
	}), "withdraw.inconsistency: failed sending tokens ")

	return nil, nil
}
