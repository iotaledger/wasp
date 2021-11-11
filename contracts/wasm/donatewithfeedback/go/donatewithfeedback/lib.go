// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// (Re-)generated by schema tool
// >>>> DO NOT CHANGE THIS FILE! <<<<
// Change the json schema instead

//nolint:dupl
package donatewithfeedback

import "github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib"

func OnLoad() {
	exports := wasmlib.NewScExports()
	exports.AddFunc(FuncDonate,       funcDonateThunk)
	exports.AddFunc(FuncWithdraw,     funcWithdrawThunk)
	exports.AddView(ViewDonation,     viewDonationThunk)
	exports.AddView(ViewDonationInfo, viewDonationInfoThunk)

	for i, key := range keyMap {
		idxMap[i] = key.KeyID()
	}
}

type DonateContext struct {
	Params  ImmutableDonateParams
	State   MutableDonateWithFeedbackState
}

func funcDonateThunk(ctx wasmlib.ScFuncContext) {
	ctx.Log("donatewithfeedback.funcDonate")
	f := &DonateContext{
		Params: ImmutableDonateParams{
			id: wasmlib.OBJ_ID_PARAMS,
		},
		State: MutableDonateWithFeedbackState{
			id: wasmlib.OBJ_ID_STATE,
		},
	}
	funcDonate(ctx, f)
	ctx.Log("donatewithfeedback.funcDonate ok")
}

type WithdrawContext struct {
	Params  ImmutableWithdrawParams
	State   MutableDonateWithFeedbackState
}

func funcWithdrawThunk(ctx wasmlib.ScFuncContext) {
	ctx.Log("donatewithfeedback.funcWithdraw")

	// only SC creator can withdraw donated funds
	ctx.Require(ctx.Caller() == ctx.ContractCreator(), "no permission")

	f := &WithdrawContext{
		Params: ImmutableWithdrawParams{
			id: wasmlib.OBJ_ID_PARAMS,
		},
		State: MutableDonateWithFeedbackState{
			id: wasmlib.OBJ_ID_STATE,
		},
	}
	funcWithdraw(ctx, f)
	ctx.Log("donatewithfeedback.funcWithdraw ok")
}

type DonationContext struct {
	Params  ImmutableDonationParams
	Results MutableDonationResults
	State   ImmutableDonateWithFeedbackState
}

func viewDonationThunk(ctx wasmlib.ScViewContext) {
	ctx.Log("donatewithfeedback.viewDonation")
	f := &DonationContext{
		Params: ImmutableDonationParams{
			id: wasmlib.OBJ_ID_PARAMS,
		},
		Results: MutableDonationResults{
			id: wasmlib.OBJ_ID_RESULTS,
		},
		State: ImmutableDonateWithFeedbackState{
			id: wasmlib.OBJ_ID_STATE,
		},
	}
	ctx.Require(f.Params.Nr().Exists(), "missing mandatory nr")
	viewDonation(ctx, f)
	ctx.Log("donatewithfeedback.viewDonation ok")
}

type DonationInfoContext struct {
	Results MutableDonationInfoResults
	State   ImmutableDonateWithFeedbackState
}

func viewDonationInfoThunk(ctx wasmlib.ScViewContext) {
	ctx.Log("donatewithfeedback.viewDonationInfo")
	f := &DonationInfoContext{
		Results: MutableDonationInfoResults{
			id: wasmlib.OBJ_ID_RESULTS,
		},
		State: ImmutableDonateWithFeedbackState{
			id: wasmlib.OBJ_ID_STATE,
		},
	}
	viewDonationInfo(ctx, f)
	ctx.Log("donatewithfeedback.viewDonationInfo ok")
}
