package sbtestsc

import (
	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/gas"
)

// withdrawFromChain withdraws all the available balance existing on the target chain
func withdrawFromChain(ctx isc.Sandbox, withdrawal coin.Value, gasReserve *uint64, gasReserveTransferAccountToChain *uint64) {
	ctx.Log().Infof(FuncWithdrawFromChain.Name)

	// if it is not already present in the SC's account the caller should have
	// provided enough base tokens to cover the gas fees for the current call
	// and for the storage deposit plus gas fees for the outgoing request to
	// accounts.transferAllowanceTo()
	ctx.TransferAllowedFunds(ctx.AccountID())

	// gasReserve is the gas fee for the 'TransferAllowanceTo' function call ub 'TransferAccountToChain'
	gasReserve = lo.CoalesceOrEmpty(gasReserve, &gas.LimitsDefault.MinGasPerRequest)
	gasReserveTransferAccountToChain = lo.CoalesceOrEmpty(gasReserveTransferAccountToChain, &gas.LimitsDefault.MinGasPerRequest)
	const storageDeposit = 20_000

	// make sure to send enough to cover the storage deposit and gas fees
	// the storage deposit will be returned along with the withdrawal
	ctx.Send(isc.RequestParameters{
		TargetAddress: ctx.ChainID().AsAddress(),
		Assets:        isc.NewAssets(coin.Value(storageDeposit + *gasReserveTransferAccountToChain + *gasReserve)),
		Metadata: &isc.SendMetadata{
			Message:   accounts.FuncTransferAllowanceTo.Message(ctx.AccountID()),
			GasBudget: *gasReserve,
			Allowance: isc.NewAssets(withdrawal + coin.Value(storageDeposit+*gasReserve)),
		},
	})

	ctx.Log().Infof("%s: success", FuncWithdrawFromChain.Name)
}
