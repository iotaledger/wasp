package sbtestsc

import (
	"math"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

// withdrawFromChain withdraws all the available balance existing on the target chain
func withdrawFromChain(ctx isc.Sandbox) dict.Dict {
	ctx.Log().Infof(FuncWithdrawFromChain.Name)
	params := kvdecoder.New(ctx.Params(), ctx.Log())
	targetChain := params.MustGetChainID(ParamChainID)
	baseTokensToWithdrawal := params.MustGetUint64(ParamBaseTokensToWithdrawal)
	availableBaseTokens := ctx.AllowanceAvailable().Assets.BaseTokens

	request := isc.RequestParameters{
		TargetAddress:  targetChain.AsAddress(),
		FungibleTokens: isc.NewFungibleBaseTokens(availableBaseTokens),
		Metadata: &isc.SendMetadata{
			TargetContract: accounts.Contract.Hname(),
			EntryPoint:     accounts.FuncWithdraw.Hname(),
			GasBudget:      math.MaxUint64,
			Allowance:      isc.NewAllowanceBaseTokens(baseTokensToWithdrawal),
		},
	}
	requiredDustDeposit := ctx.EstimateRequiredDustDeposit(request)
	if availableBaseTokens < requiredDustDeposit {
		ctx.Log().Panicf("not enough base tokens sent to cover dust deposit")
	}
	ctx.TransferAllowedFunds(ctx.AccountID())
	ctx.Send(request)

	ctx.Log().Infof("%s: success", FuncWithdrawFromChain)
	return nil
}
