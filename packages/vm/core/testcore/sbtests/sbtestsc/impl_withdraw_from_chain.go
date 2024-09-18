package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

// withdrawFromChain withdraws all the available balance existing on the target chain
func withdrawFromChain(ctx isc.Sandbox) isc.CallArguments {
	ctx.Log().Infof(FuncWithdrawFromChain.Name)
	params := ctx.Params()
	targetChain := params.MustGetChainID(ParamChainID)
	withdrawal := params.MustGetUint64(ParamBaseTokens)

	// if it is not already present in the SC's account the caller should have
	// provided enough base tokens to cover the gas fees for the current call
	// and for the storage deposit plus gas fees for the outgoing request to
	// accounts.transferAllowanceTo()
	ctx.TransferAllowedFunds(ctx.AccountID())

	// gasReserve is the gas fee for the 'TransferAllowanceTo' function call ub 'TransferAccountToChain'
	gasReserve := params.MustGetUint64(ParamGasReserve, gas.LimitsDefault.MinGasPerRequest)
	gasReserveTransferAccountToChain := params.MustGetUint64(ParamGasReserveTransferAccountToChain, gas.LimitsDefault.MinGasPerRequest)
	const storageDeposit = 20_000

	// make sure to send enough to cover the storage deposit and gas fees
	// the storage deposit will be returned along with the withdrawal
	ctx.Send(isc.RequestParameters{
		TargetAddress: targetChain.AsAddress(),
		Assets:        isc.NewAssets(storageDeposit + gasReserveTransferAccountToChain + gasReserve),
		Metadata: &isc.SendMetadata{
			Message:   accounts.FuncTransferAccountToChain.Message(&gasReserve),
			GasBudget: gasReserve,
			Allowance: isc.NewAssets(withdrawal + storageDeposit + gasReserve),
		},
	})

	ctx.Log().Infof("%s: success", FuncWithdrawFromChain.Name)
	return nil
}
