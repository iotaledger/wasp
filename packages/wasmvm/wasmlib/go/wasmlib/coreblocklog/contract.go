// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// (Re-)generated by schema tool
// >>>> DO NOT CHANGE THIS FILE! <<<<
// Change the schema definition file instead

package coreblocklog

import "github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"

type ControlAddressesCall struct {
	Func    *wasmlib.ScView
	Results ImmutableControlAddressesResults
}

type GetBlockInfoCall struct {
	Func    *wasmlib.ScView
	Params  MutableGetBlockInfoParams
	Results ImmutableGetBlockInfoResults
}

type GetEventsForBlockCall struct {
	Func    *wasmlib.ScView
	Params  MutableGetEventsForBlockParams
	Results ImmutableGetEventsForBlockResults
}

type GetEventsForContractCall struct {
	Func    *wasmlib.ScView
	Params  MutableGetEventsForContractParams
	Results ImmutableGetEventsForContractResults
}

type GetEventsForRequestCall struct {
	Func    *wasmlib.ScView
	Params  MutableGetEventsForRequestParams
	Results ImmutableGetEventsForRequestResults
}

type GetRequestIDsForBlockCall struct {
	Func    *wasmlib.ScView
	Params  MutableGetRequestIDsForBlockParams
	Results ImmutableGetRequestIDsForBlockResults
}

type GetRequestReceiptCall struct {
	Func    *wasmlib.ScView
	Params  MutableGetRequestReceiptParams
	Results ImmutableGetRequestReceiptResults
}

type GetRequestReceiptsForBlockCall struct {
	Func    *wasmlib.ScView
	Params  MutableGetRequestReceiptsForBlockParams
	Results ImmutableGetRequestReceiptsForBlockResults
}

type IsRequestProcessedCall struct {
	Func    *wasmlib.ScView
	Params  MutableIsRequestProcessedParams
	Results ImmutableIsRequestProcessedResults
}

type Funcs struct{}

var ScFuncs Funcs

func (sc Funcs) ControlAddresses(ctx wasmlib.ScViewCallContext) *ControlAddressesCall {
	f := &ControlAddressesCall{Func: wasmlib.NewScView(ctx, HScName, HViewControlAddresses)}
	wasmlib.NewCallResultsProxy(f.Func, &f.Results.proxy)
	return f
}

func (sc Funcs) GetBlockInfo(ctx wasmlib.ScViewCallContext) *GetBlockInfoCall {
	f := &GetBlockInfoCall{Func: wasmlib.NewScView(ctx, HScName, HViewGetBlockInfo)}
	f.Params.proxy = wasmlib.NewCallParamsProxy(f.Func)
	wasmlib.NewCallResultsProxy(f.Func, &f.Results.proxy)
	return f
}

func (sc Funcs) GetEventsForBlock(ctx wasmlib.ScViewCallContext) *GetEventsForBlockCall {
	f := &GetEventsForBlockCall{Func: wasmlib.NewScView(ctx, HScName, HViewGetEventsForBlock)}
	f.Params.proxy = wasmlib.NewCallParamsProxy(f.Func)
	wasmlib.NewCallResultsProxy(f.Func, &f.Results.proxy)
	return f
}

func (sc Funcs) GetEventsForContract(ctx wasmlib.ScViewCallContext) *GetEventsForContractCall {
	f := &GetEventsForContractCall{Func: wasmlib.NewScView(ctx, HScName, HViewGetEventsForContract)}
	f.Params.proxy = wasmlib.NewCallParamsProxy(f.Func)
	wasmlib.NewCallResultsProxy(f.Func, &f.Results.proxy)
	return f
}

func (sc Funcs) GetEventsForRequest(ctx wasmlib.ScViewCallContext) *GetEventsForRequestCall {
	f := &GetEventsForRequestCall{Func: wasmlib.NewScView(ctx, HScName, HViewGetEventsForRequest)}
	f.Params.proxy = wasmlib.NewCallParamsProxy(f.Func)
	wasmlib.NewCallResultsProxy(f.Func, &f.Results.proxy)
	return f
}

func (sc Funcs) GetRequestIDsForBlock(ctx wasmlib.ScViewCallContext) *GetRequestIDsForBlockCall {
	f := &GetRequestIDsForBlockCall{Func: wasmlib.NewScView(ctx, HScName, HViewGetRequestIDsForBlock)}
	f.Params.proxy = wasmlib.NewCallParamsProxy(f.Func)
	wasmlib.NewCallResultsProxy(f.Func, &f.Results.proxy)
	return f
}

func (sc Funcs) GetRequestReceipt(ctx wasmlib.ScViewCallContext) *GetRequestReceiptCall {
	f := &GetRequestReceiptCall{Func: wasmlib.NewScView(ctx, HScName, HViewGetRequestReceipt)}
	f.Params.proxy = wasmlib.NewCallParamsProxy(f.Func)
	wasmlib.NewCallResultsProxy(f.Func, &f.Results.proxy)
	return f
}

func (sc Funcs) GetRequestReceiptsForBlock(ctx wasmlib.ScViewCallContext) *GetRequestReceiptsForBlockCall {
	f := &GetRequestReceiptsForBlockCall{Func: wasmlib.NewScView(ctx, HScName, HViewGetRequestReceiptsForBlock)}
	f.Params.proxy = wasmlib.NewCallParamsProxy(f.Func)
	wasmlib.NewCallResultsProxy(f.Func, &f.Results.proxy)
	return f
}

func (sc Funcs) IsRequestProcessed(ctx wasmlib.ScViewCallContext) *IsRequestProcessedCall {
	f := &IsRequestProcessedCall{Func: wasmlib.NewScView(ctx, HScName, HViewIsRequestProcessed)}
	f.Params.proxy = wasmlib.NewCallParamsProxy(f.Func)
	wasmlib.NewCallResultsProxy(f.Func, &f.Results.proxy)
	return f
}

var exportMap = wasmlib.ScExportMap{
	Names: []string{
		ViewControlAddresses,
		ViewGetBlockInfo,
		ViewGetEventsForBlock,
		ViewGetEventsForContract,
		ViewGetEventsForRequest,
		ViewGetRequestIDsForBlock,
		ViewGetRequestReceipt,
		ViewGetRequestReceiptsForBlock,
		ViewIsRequestProcessed,
	},
	Funcs: []wasmlib.ScFuncContextFunction{
	},
	Views: []wasmlib.ScViewContextFunction{
		wasmlib.ViewError,
		wasmlib.ViewError,
		wasmlib.ViewError,
		wasmlib.ViewError,
		wasmlib.ViewError,
		wasmlib.ViewError,
		wasmlib.ViewError,
		wasmlib.ViewError,
		wasmlib.ViewError,
	},
}

func OnDispatch(index int32) {
	if index == -1 {
		exportMap.Export()
		return
	}

	panic("Calling core contract?")
}
