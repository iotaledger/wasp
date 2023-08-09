package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
)

// withdrawFromChain withdraws all the available balance existing on the target chain
func withdrawFromChain(ctx isc.Sandbox) dict.Dict {
	ctx.Log().Infof(FuncWithdrawFromChain.Name)
	params := ctx.Params()
	targetChain := params.MustGetChainID(ParamChainID)
	withdrawal := params.MustGetUint64(ParamBaseTokens)

	// if it is not already present in the SC's account the caller should have
	// provided enough base tokens to cover the gas fees for the current call
	// (should be wasmlib.MinGasFee in default), and for the storage deposit
	// plus gas fees for the outgoing request to accounts.transferAllowanceTo()
	ctx.TransferAllowedFunds(ctx.AccountID())

	// gasReserve is the gas fee for the 'TransferAllowanceTo' function call ub 'TransferAccountToChain'
	gasReserve := params.MustGetUint64(ParamGasReserve, gas.LimitsDefault.MinGasPerRequest)
	gasReserveTransferAccountToChain := params.MustGetUint64(ParamGasReserveTransferAccountToChain, gas.LimitsDefault.MinGasPerRequest)
	const storageDeposit = wasmlib.StorageDeposit

	// make sure to send enough to cover the storage deposit and gas fees
	// the storage deposit will be returned along with the withdrawal
	ctx.Send(isc.RequestParameters{
		TargetAddress: targetChain.AsAddress(),
		Assets:        isc.NewAssetsBaseTokens(storageDeposit + gasReserveTransferAccountToChain + gasReserve),
		Metadata: &isc.SendMetadata{
			TargetContract: accounts.Contract.Hname(),
			EntryPoint:     accounts.FuncTransferAccountToChain.Hname(),
			Params: dict.Dict{
				accounts.ParamGasReserve: codec.EncodeUint64(gasReserve),
			},
			GasBudget: gasReserve,
			Allowance: isc.NewAssetsBaseTokens(withdrawal + storageDeposit + gasReserve),
		},
	})

	ctx.Log().Infof("%s: success", FuncWithdrawFromChain.Name)
	return nil
}
