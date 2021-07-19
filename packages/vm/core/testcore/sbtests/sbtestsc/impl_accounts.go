package sbtestsc

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

// calls withdrawToChain to the chain ID
func withdrawToChain(ctx iscp.Sandbox) (dict.Dict, error) {
	ctx.Log().Infof(FuncWithdrawToChain.Name)
	params := kvdecoder.New(ctx.Params(), ctx.Log())
	targetChain := params.MustGetChainID(ParamChainID)
	succ := ctx.Send(targetChain.AsAddress(), iscp.NewTransferIotas(1), &iscp.SendMetadata{
		TargetContract: accounts.Contract.Hname(),
		EntryPoint:     accounts.FuncWithdraw.Hname(),
		Args:           nil,
	})
	if !succ {
		return nil, fmt.Errorf("failed to post request")
	}
	ctx.Log().Infof("%s: success", FuncWithdrawToChain)
	return nil, nil
}
