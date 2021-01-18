package test_sandbox_sc

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

// calls withdrawToChain to the chain ID
func withdrawToChain(ctx vmtypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Infof(FuncWithdrawToChain)
	targetChain, ok, err := codec.DecodeChainID(ctx.Params().MustGet(ParamChainID))
	if err != nil || !ok {
		ctx.Log().Panicf("wrong parameter '%s'", ParamChainID)
	}
	succ := ctx.PostRequest(vmtypes.PostRequestParams{
		TargetContractID: accounts.Interface.ContractID(targetChain),
		EntryPoint:       coretypes.Hn(accounts.FuncWithdrawToChain),
		Transfer: cbalances.NewFromMap(map[balance.Color]int64{
			balance.ColorIOTA: 2,
		}),
	})
	if !succ {
		return nil, fmt.Errorf("failed to post request")
	}
	ctx.Log().Infof("%s: success", FuncWithdrawToChain)
	return nil, nil
}
