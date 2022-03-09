// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package evmimpl

import (
	"math"
	"time"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/assert"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
)

func setBlockTime(ctx iscp.Sandbox) dict.Dict {
	ctx.RequireCallerIsChainOwner()

	params := kvdecoder.New(ctx.Params(), ctx.Log())
	a := assert.NewAssert(ctx.Log())

	blockTime := params.MustGetUint32(evm.FieldBlockTime)
	a.Requiref(blockTime > 0, "blockTime must be > 0")

	mustSchedule := !ctx.State().MustHas(keyBlockTime)

	ctx.State().Set(keyBlockTime, codec.EncodeUint32(blockTime))
	if mustSchedule {
		scheduleNextBlock(ctx)
	}
	return nil
}

func getBlockTime(state kv.KVStoreReader) uint32 {
	bt, _ := codec.DecodeUint32(state.MustGet(keyBlockTime), 0)
	return bt
}

func scheduleNextBlock(ctx iscp.Sandbox) {
	ctx.RequireCallerAnyOf([]*iscp.AgentID{ctx.ChainOwnerID(), ctx.ContractAgentID()})

	blockTime := getBlockTime(ctx.State())
	if blockTime == 0 {
		return
	}

	ctx.Send(iscp.RequestParameters{
		TargetAddress:              ctx.ChainID().AsAddress(),
		FungibleTokens:             iscp.NewFungibleTokens(1, nil),
		AdjustToMinimumDustDeposit: true,
		Metadata: &iscp.SendMetadata{
			TargetContract: ctx.Contract(),
			EntryPoint:     evm.FuncMintBlock.Hname(),
			GasBudget:      math.MaxUint64, // TODO: ?
		},
		Options: iscp.SendOptions{Timelock: &iscp.TimeData{
			Time: time.Unix(0, ctx.Timestamp()).
				Add(time.Duration(blockTime) * time.Second),
		}},
	})
}
