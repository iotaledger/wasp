// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"github.com/iotaledger/wasp/contracts/native/evm"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
)

func requireOwner(ctx iscp.Sandbox, allowSelf ...bool) {
	contractOwner, err := codec.DecodeAgentID(ctx.State().MustGet(keyEVMOwner))
	ctx.RequireNoError(err)

	allowed := []*iscp.AgentID{contractOwner}
	if len(allowSelf) > 0 && allowSelf[0] {
		allowed = append(allowed, iscp.NewAgentID(ctx.ChainID().AsAddress(), ctx.Contract()))
	}

	ctx.RequireCallerAnyOf(allowed)
}

func setNextOwner(ctx iscp.Sandbox) dict.Dict {
	requireOwner(ctx)
	par := kvdecoder.New(ctx.Params(), ctx.Log())
	ctx.State().Set(keyNextEVMOwner, codec.EncodeAgentID(par.MustGetAgentID(evm.FieldNextEVMOwner)))
	return nil
}

func claimOwnership(ctx iscp.Sandbox) dict.Dict {
	nextOwner, err := codec.DecodeAgentID(ctx.State().MustGet(keyNextEVMOwner))
	ctx.RequireNoError(err)
	ctx.RequireCaller(nextOwner)

	ctx.State().Set(keyEVMOwner, codec.EncodeAgentID(nextOwner))
	scheduleNextBlock(ctx)
	return nil
}

func getOwner(ctx iscp.SandboxView) dict.Dict {
	return result(ctx.State().MustGet(keyEVMOwner))
}
