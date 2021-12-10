// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evminternal

import (
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/iotaledger/wasp/contracts/native/evm"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/assert"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
)

const (
	keyGasPerIota   = "g"
	keyEVMOwner     = "o"
	keyNextEVMOwner = "n"
	keyBlockTime    = "b"

	// keyEVMState is the subrealm prefix for the EVM state
	keyEVMState = "s"
)

var ManagementHandlers = []coreutil.ProcessorEntryPoint{
	evm.FuncSetNextOwner.WithHandler(setNextOwner),
	evm.FuncClaimOwnership.WithHandler(claimOwnership),
	evm.FuncSetGasPerIota.WithHandler(setGasPerIota),
	evm.FuncWithdrawGasFees.WithHandler(withdrawGasFees),
	evm.FuncGetOwner.WithHandler(getOwner),
	evm.FuncGetGasPerIota.WithHandler(getGasPerIota),
	evm.FuncSetBlockTime.WithHandler(setBlockTime),
}

func EVMStateSubrealm(state kv.KVStore) kv.KVStore {
	return subrealm.New(state, keyEVMState)
}

func InitializeManagement(ctx iscp.Sandbox) {
	ctx.State().Set(keyGasPerIota, codec.EncodeUint64(evm.DefaultGasPerIota))
	ctx.State().Set(keyEVMOwner, codec.EncodeAgentID(ctx.ContractCreator()))
}

func setBlockTime(ctx iscp.Sandbox) (dict.Dict, error) {
	requireOwner(ctx)

	params := kvdecoder.New(ctx.Params(), ctx.Log())
	a := assert.NewAssert(ctx.Log())

	blockTime := params.MustGetUint32(evm.FieldBlockTime)
	a.Require(blockTime > 0, "blockTime must be > 0")

	mustSchedule := !ctx.State().MustHas(keyBlockTime)

	ctx.State().Set(keyBlockTime, codec.EncodeUint32(blockTime))
	if mustSchedule {
		ScheduleNextBlock(ctx)
	}
	return nil, nil
}

func getBlockTime(state kv.KVStoreReader) uint32 {
	bt, _ := codec.DecodeUint32(state.MustGet(keyBlockTime), 0)
	return bt
}

func ScheduleNextBlock(ctx iscp.Sandbox) {
	requireOwner(ctx, true)

	a := assert.NewAssert(ctx.Log())

	blockTime := getBlockTime(ctx.State())
	a.Require(blockTime > 0, "ScheduleNextBlock: blockTime must be > 0")

	ok := ctx.Send(ctx.ChainID().AsAddress(), colored.NewBalancesForIotas(1), &iscp.SendMetadata{
		TargetContract: ctx.Contract(),
		EntryPoint:     evm.FuncMintBlock.Hname(),
	}, iscp.SendOptions{
		TimeLock: uint32(time.Unix(0, ctx.GetTimestamp()).Unix()) + blockTime,
	})
	a.Require(ok, "failed to schedule next block")
}

func requireOwner(ctx iscp.Sandbox, allowSelf ...bool) {
	contractOwner, err := codec.DecodeAgentID(ctx.State().MustGet(keyEVMOwner))
	a := assert.NewAssert(ctx.Log())
	a.RequireNoError(err)

	allowed := []*iscp.AgentID{contractOwner}
	if len(allowSelf) > 0 && allowSelf[0] {
		allowed = append(allowed, iscp.NewAgentID(ctx.ChainID().AsAddress(), ctx.Contract()))
	}

	a.RequireCaller(ctx, allowed)
}

func setNextOwner(ctx iscp.Sandbox) (dict.Dict, error) {
	requireOwner(ctx)
	par := kvdecoder.New(ctx.Params(), ctx.Log())
	ctx.State().Set(keyNextEVMOwner, codec.EncodeAgentID(par.MustGetAgentID(evm.FieldNextEVMOwner)))
	return nil, nil
}

func claimOwnership(ctx iscp.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())

	nextOwner, err := codec.DecodeAgentID(ctx.State().MustGet(keyNextEVMOwner))
	a.RequireNoError(err)
	a.RequireCaller(ctx, []*iscp.AgentID{nextOwner})

	ctx.State().Set(keyEVMOwner, codec.EncodeAgentID(nextOwner))
	return nil, nil
}

func getOwner(ctx iscp.SandboxView) (dict.Dict, error) {
	return Result(ctx.State().MustGet(keyEVMOwner)), nil
}

func setGasPerIota(ctx iscp.Sandbox) (dict.Dict, error) {
	requireOwner(ctx)
	par := kvdecoder.New(ctx.Params())
	gasPerIotaBin := codec.EncodeUint64(par.MustGetUint64(evm.FieldGasPerIota))
	ctx.State().Set(keyGasPerIota, gasPerIotaBin)
	return nil, nil
}

func getGasPerIota(ctx iscp.SandboxView) (dict.Dict, error) {
	return Result(ctx.State().MustGet(keyGasPerIota)), nil
}

func withdrawGasFees(ctx iscp.Sandbox) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())
	requireOwner(ctx)

	paramsDecoder := kvdecoder.New(ctx.Params(), ctx.Log())
	targetAgentID := paramsDecoder.MustGetAgentID(evm.FieldAgentID, ctx.Caller())

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

func ApplyTransaction(ctx iscp.Sandbox, apply func(tx *types.Transaction, blockTime uint32) (*types.Receipt, error)) (dict.Dict, error) {
	a := assert.NewAssert(ctx.Log())

	tx := &types.Transaction{}
	err := tx.UnmarshalBinary(ctx.Params().MustGet(evm.FieldTransactionData))
	a.RequireNoError(err)

	transferredIotas, gasPerIota := takeGasFee(ctx, tx)

	blockTime := getBlockTime(ctx.State())
	receipt, err := apply(tx, blockTime)
	a.RequireNoError(err)

	return refundUnusedGasFee(ctx, ctx.Caller(), transferredIotas, gasPerIota, receipt.GasUsed), nil
}

func takeGasFee(ctx iscp.Sandbox, tx *types.Transaction) (uint64, uint64) {
	a := assert.NewAssert(ctx.Log())

	transferredIotas := ctx.IncomingTransfer().Get(getFeeColor(ctx))
	gasPerIota, err := codec.DecodeUint64(ctx.State().MustGet(keyGasPerIota), 0)
	a.RequireNoError(err)
	txGasLimit := tx.Gas()

	a.Require(
		transferredIotas >= txGasLimit/gasPerIota,
		"transferred tokens (%d) not enough to cover the gas limit set in the transaction (%d at %d gas per iota token)", transferredIotas, txGasLimit, gasPerIota,
	)

	return transferredIotas, gasPerIota
}

func refundUnusedGasFee(ctx iscp.Sandbox, caller *iscp.AgentID, transferredIotas, gasPerIota, gasUsed uint64) dict.Dict {
	iotasGasFee := gasUsed / gasPerIota
	if transferredIotas > iotasGasFee {
		// refund unspent gas fee to the sender's on-chain account
		iotasGasRefund := transferredIotas - iotasGasFee
		_, err := ctx.Call(
			accounts.Contract.Hname(),
			accounts.FuncDeposit.Hname(),
			dict.Dict{accounts.ParamAgentID: codec.EncodeAgentID(caller)},
			colored.NewBalancesForIotas(iotasGasRefund),
		)
		a := assert.NewAssert(ctx.Log())
		a.RequireNoError(err)
	}

	return dict.Dict{
		evm.FieldGasFee:  codec.EncodeUint64(iotasGasFee),
		evm.FieldGasUsed: codec.EncodeUint64(gasUsed),
	}
}

func getFeeColor(ctx iscp.Sandbox) colored.Color {
	a := assert.NewAssert(ctx.Log())

	// call root contract view to get the feecolor
	feeInfo, err := ctx.Call(
		governance.Contract.Hname(),
		governance.FuncGetFeeInfo.Hname(),
		dict.Dict{governance.ParamHname: ctx.Contract().Bytes()},
		nil,
	)
	a.RequireNoError(err)
	feeColor, err := codec.DecodeColor(feeInfo.MustGet(governance.ParamFeeColor))
	a.RequireNoError(err)
	return feeColor
}
