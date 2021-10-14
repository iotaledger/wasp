// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evminternal

import (
	"github.com/iotaledger/wasp/contracts/native/evm"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/assert"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
)

const (
	keyGasPerIota   = "g"
	keyEVMOwner     = "o"
	keyNextEVMOwner = "n"
)

var ManagementHandlers = []coreutil.ProcessorEntryPoint{
	evm.FuncSetNextOwner.WithHandler(setNextOwner),
	evm.FuncClaimOwnership.WithHandler(claimOwnership),
	evm.FuncSetGasPerIota.WithHandler(setGasPerIota),
	evm.FuncWithdrawGasFees.WithHandler(withdrawGasFees),
	evm.FuncGetOwner.WithHandler(getOwner),
	evm.FuncGetGasPerIota.WithHandler(getGasPerIota),
}

func InitializeManagement(ctx iscp.Sandbox) {
	ctx.State().Set(keyGasPerIota, codec.EncodeUint64(evm.DefaultGasPerIota))
	ctx.State().Set(keyEVMOwner, codec.EncodeAgentID(ctx.ContractCreator()))
}

func requireOwner(ctx iscp.Sandbox) {
	contractOwner, err := codec.DecodeAgentID(ctx.State().MustGet(keyEVMOwner))
	a := assert.NewAssert(ctx.Log())
	a.RequireNoError(err)
	a.Require(contractOwner.Equals(ctx.Caller()), "can only be called by the contract owner")
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
	a.Require(nextOwner.Equals(ctx.Caller()), "Can only be called by the contract owner")

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

func RequireGasFee(ctx iscp.Sandbox, txGasLimit uint64, f func() uint64) dict.Dict {
	a := assert.NewAssert(ctx.Log())

	transferredIotas := ctx.IncomingTransfer().Get(getFeeColor(ctx))
	gasPerIota, err := codec.DecodeUint64(ctx.State().MustGet(keyGasPerIota), 0)
	a.RequireNoError(err)

	a.Require(
		transferredIotas >= txGasLimit/gasPerIota,
		"transferred tokens (%d) not enough to cover the gas limit set in the transaction (%d at %d gas per iota token)", transferredIotas, txGasLimit, gasPerIota,
	)

	gasUsed := f()

	iotasGasFee := gasUsed / gasPerIota
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
