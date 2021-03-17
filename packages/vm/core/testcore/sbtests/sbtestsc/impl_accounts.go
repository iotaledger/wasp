package sbtestsc

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

// calls withdrawToChain to the chain ID
func withdrawToChain(ctx coretypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Infof(FuncWithdrawToChain)
	params := kvdecoder.New(ctx.Params(), ctx.Log())
	targetChain := params.MustGetChainID(ParamChainID)
	succ := ctx.PostRequest(coretypes.PostRequestParams{
		TargetContractID: accounts.Interface.ContractID(*targetChain),
		EntryPoint:       coretypes.Hn(accounts.FuncWithdrawToChain),
		Transfer:         coretypes.NewTransferIotas(2),
	})
	if !succ {
		return nil, fmt.Errorf("failed to post request")
	}
	ctx.Log().Infof("%s: success", FuncWithdrawToChain)
	return nil, nil
}
