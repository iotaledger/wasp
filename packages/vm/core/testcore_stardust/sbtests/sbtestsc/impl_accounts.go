package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// calls withdrawToChain to the chain ID
func withdrawToChain(ctx iscp.Sandbox) dict.Dict {
	ctx.Log().Infof(FuncWithdrawToChain.Name)
	//params := kvdecoder.New(ctx.Params(), ctx.Log())
	//targetChain := params.MustGetChainID(ParamChainID)
	//succ := ctx.Send(targetChain.AsAddress(), iscp.NewAssets(1, nil), &iscp.SendMetadata{
	//	TargetContract: accounts.Contract.Hname(),
	//	EntryPoint:     accounts.FuncWithdraw.Hname(),
	//	Params:         nil,
	//})
	//if !succ {
	//	return nil, fmt.Errorf("failed to post request")
	//}
	//ctx.Log().Infof("%s: success", FuncWithdrawToChain)
	return nil
}
