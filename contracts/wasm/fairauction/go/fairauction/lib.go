// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// (Re-)generated by schema tool
// >>>> DO NOT CHANGE THIS FILE! <<<<
// Change the json schema instead

//nolint:dupl
package fairauction

import "github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib"

func OnLoad() {
	exports := wasmlib.NewScExports()
	exports.AddFunc(FuncFinalizeAuction, funcFinalizeAuctionThunk)
	exports.AddFunc(FuncPlaceBid,        funcPlaceBidThunk)
	exports.AddFunc(FuncSetOwnerMargin,  funcSetOwnerMarginThunk)
	exports.AddFunc(FuncStartAuction,    funcStartAuctionThunk)
	exports.AddView(ViewGetInfo,         viewGetInfoThunk)

	for i, key := range keyMap {
		idxMap[i] = key.KeyID()
	}
}

type FinalizeAuctionContext struct {
	Params  ImmutableFinalizeAuctionParams
	State   MutableFairAuctionState
}

func funcFinalizeAuctionThunk(ctx wasmlib.ScFuncContext) {
	ctx.Log("fairauction.funcFinalizeAuction")

	// only SC itself can invoke this function
	ctx.Require(ctx.Caller() == ctx.AccountID(), "no permission")

	f := &FinalizeAuctionContext{
		Params: ImmutableFinalizeAuctionParams{
			id: wasmlib.OBJ_ID_PARAMS,
		},
		State: MutableFairAuctionState{
			id: wasmlib.OBJ_ID_STATE,
		},
	}
	ctx.Require(f.Params.Color().Exists(), "missing mandatory color")
	funcFinalizeAuction(ctx, f)
	ctx.Log("fairauction.funcFinalizeAuction ok")
}

type PlaceBidContext struct {
	Params  ImmutablePlaceBidParams
	State   MutableFairAuctionState
}

func funcPlaceBidThunk(ctx wasmlib.ScFuncContext) {
	ctx.Log("fairauction.funcPlaceBid")
	f := &PlaceBidContext{
		Params: ImmutablePlaceBidParams{
			id: wasmlib.OBJ_ID_PARAMS,
		},
		State: MutableFairAuctionState{
			id: wasmlib.OBJ_ID_STATE,
		},
	}
	ctx.Require(f.Params.Color().Exists(), "missing mandatory color")
	funcPlaceBid(ctx, f)
	ctx.Log("fairauction.funcPlaceBid ok")
}

type SetOwnerMarginContext struct {
	Params  ImmutableSetOwnerMarginParams
	State   MutableFairAuctionState
}

func funcSetOwnerMarginThunk(ctx wasmlib.ScFuncContext) {
	ctx.Log("fairauction.funcSetOwnerMargin")

	// only SC creator can set owner margin
	ctx.Require(ctx.Caller() == ctx.ContractCreator(), "no permission")

	f := &SetOwnerMarginContext{
		Params: ImmutableSetOwnerMarginParams{
			id: wasmlib.OBJ_ID_PARAMS,
		},
		State: MutableFairAuctionState{
			id: wasmlib.OBJ_ID_STATE,
		},
	}
	ctx.Require(f.Params.OwnerMargin().Exists(), "missing mandatory ownerMargin")
	funcSetOwnerMargin(ctx, f)
	ctx.Log("fairauction.funcSetOwnerMargin ok")
}

type StartAuctionContext struct {
	Params  ImmutableStartAuctionParams
	State   MutableFairAuctionState
}

func funcStartAuctionThunk(ctx wasmlib.ScFuncContext) {
	ctx.Log("fairauction.funcStartAuction")
	f := &StartAuctionContext{
		Params: ImmutableStartAuctionParams{
			id: wasmlib.OBJ_ID_PARAMS,
		},
		State: MutableFairAuctionState{
			id: wasmlib.OBJ_ID_STATE,
		},
	}
	ctx.Require(f.Params.Color().Exists(), "missing mandatory color")
	ctx.Require(f.Params.MinimumBid().Exists(), "missing mandatory minimumBid")
	funcStartAuction(ctx, f)
	ctx.Log("fairauction.funcStartAuction ok")
}

type GetInfoContext struct {
	Params  ImmutableGetInfoParams
	Results MutableGetInfoResults
	State   ImmutableFairAuctionState
}

func viewGetInfoThunk(ctx wasmlib.ScViewContext) {
	ctx.Log("fairauction.viewGetInfo")
	f := &GetInfoContext{
		Params: ImmutableGetInfoParams{
			id: wasmlib.OBJ_ID_PARAMS,
		},
		Results: MutableGetInfoResults{
			id: wasmlib.OBJ_ID_RESULTS,
		},
		State: ImmutableFairAuctionState{
			id: wasmlib.OBJ_ID_STATE,
		},
	}
	ctx.Require(f.Params.Color().Exists(), "missing mandatory color")
	viewGetInfo(ctx, f)
	ctx.Log("fairauction.viewGetInfo ok")
}
