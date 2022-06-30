package sbtestsc

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

type blockCtx struct {
	numCalls uint8
}

func openBlockContext(ctx iscp.Sandbox) dict.Dict {
	ctx.RequireCallerIsChainOwner()
	ctx.Privileged().SetBlockContext(&blockCtx{})
	return nil
}

func closeBlockContext(ctx iscp.Sandbox) dict.Dict {
	ctx.RequireCallerIsChainOwner()
	ctx.State().Set("numCalls", codec.EncodeUint8(getBlockContext(ctx).numCalls))
	return nil
}

func getBlockContext(ctx iscp.Sandbox) *blockCtx {
	return ctx.Privileged().BlockContext().(*blockCtx)
}

func getLastBlockNumCalls(ctx iscp.SandboxView) dict.Dict {
	return dict.Dict{"numCalls": ctx.State().MustGet("numCalls")}
}
