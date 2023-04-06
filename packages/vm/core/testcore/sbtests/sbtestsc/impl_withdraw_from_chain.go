package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
)

// withdrawFromChain withdraws all the available balance existing on the target chain
func withdrawFromChain(ctx isc.Sandbox) dict.Dict {
	ctx.Log().Infof(FuncWithdrawFromChain.Name)
	params := ctx.Params()
	targetChain := params.MustGetChainID(ParamChainID)
	withdrawal := params.MustGetUint64(ParamBaseTokens)

	// if it is not already present in the SC's account the caller should have
	// provided enough base tokens to cover the gas fees for the current call,
	// and for the storage deposit plus gas fees for the outgoing request to
	// accounts.transferAllowanceTo()
	ctx.TransferAllowedFunds(ctx.AccountID())

	// This is just a test contract, but normally these numbers should
	// be parameters because there is no way for the contract to figure
	// out the gas fees on the other chain, and it's also silly to run
	// the costly calculation to determine storage deposit every time
	// unless absolutely necessary. Better to just make sure that the
	// storage deposit is large enough, since it will be returned anyway.
	const gasFee = wasmlib.MinGasFee
	const storageDeposit = wasmlib.StorageDeposit

	// make sure to send enough to cover the storage deposit and gas fees
	// the storage deposit will be returned along with the withdrawal
	ctx.Send(isc.RequestParameters{
		TargetAddress: targetChain.AsAddress(),
		Assets:        isc.NewAssetsBaseTokens(storageDeposit + gasFee + gasFee),
		Metadata: &isc.SendMetadata{
			TargetContract: accounts.Contract.Hname(),
			EntryPoint:     accounts.FuncTransferAccountToChain.Hname(),
			GasBudget:      gasFee,
			Allowance:      isc.NewAssetsBaseTokens(withdrawal + storageDeposit + gasFee),
		},
	})

	ctx.Log().Infof("%s: success", FuncWithdrawFromChain.Name)
	return nil
}
