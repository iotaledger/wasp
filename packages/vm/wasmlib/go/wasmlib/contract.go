// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmlib

import (
	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib/wasmrequests"
	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib/wasmtypes"
)

type ScFuncCallContext interface {
	ChainID() wasmtypes.ScChainID
	InitFuncCallContext()
}

type ScViewCallContext interface {
	InitViewCallContext()
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScView struct {
	hContract    wasmtypes.ScHname
	hFunction    wasmtypes.ScHname
	params       ScDict
	resultsProxy *wasmtypes.Proxy
}

func NewScView(ctx ScViewCallContext, hContract, hFunction wasmtypes.ScHname) *ScView {
	ctx.InitViewCallContext()
	v := new(ScView)
	v.initView(hContract, hFunction)
	return v
}

func NewCallParamsProxy(v *ScView) wasmtypes.Proxy {
	v.params = NewScDict()
	return wasmtypes.NewProxy(v.params)
}

func NewCallResultsProxy(v *ScView, resultsProxy *wasmtypes.Proxy) {
	v.resultsProxy = resultsProxy
}

func (v *ScView) Call() {
	v.call(nil)
}

func (v *ScView) call(transfer ScAssets) {
	req := wasmrequests.CallRequest{
		Contract: v.hContract,
		Function: v.hFunction,
		Params:   v.params.Bytes(),
		Transfer: transfer.Bytes(),
	}
	res := Sandbox(FnCall, req.Bytes())
	if v.resultsProxy != nil {
		*v.resultsProxy = wasmtypes.NewProxy(NewScDictFromBytes(res))
	}
}

func (v *ScView) initView(hContract, hFunction wasmtypes.ScHname) {
	v.hContract = hContract
	v.hFunction = hFunction
}

func (v *ScView) OfContract(hContract wasmtypes.ScHname) *ScView {
	v.hContract = hContract
	return v
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScInitFunc struct {
	ScView
	host ScHost
}

func NewScInitFunc(ctx ScFuncCallContext, hContract, hFunction wasmtypes.ScHname) *ScInitFunc {
	f := new(ScInitFunc)
	f.initView(hContract, hFunction)
	if ctx != nil {
		ctx.InitFuncCallContext()
		return f
	}

	// Special initialization for SoloContext usage
	// Note that we do not have a contract context that can talk to the host
	// until *after* deployment of the contract, so we cannot use the normal
	// params proxy to pass parameters because it does not exist yet.
	// Instead, we use a special temporary host implementation that knows
	// just enough to gather the parameter data and pass it correctly to
	// solo's contract deployment function, which in turn passes it to the
	// contract's init() function
	f.host = ConnectHost(NewInitHost())
	return f
}

func (f *ScInitFunc) Call() {
	Panic("cannot call init")
}

func (f *ScInitFunc) OfContract(hContract wasmtypes.ScHname) *ScInitFunc {
	f.hContract = hContract
	return f
}

func (f *ScInitFunc) Params() []interface{} {
	var params []interface{}
	for k, v := range f.params {
		params = append(params, k)
		params = append(params, v)
	}
	ConnectHost(f.host)
	return params
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScFunc struct {
	ScView
	ctx      ScFuncCallContext
	delay    uint32
	transfer ScAssets
}

func NewScFunc(ctx ScFuncCallContext, hContract, hFunction wasmtypes.ScHname) *ScFunc {
	ctx.InitFuncCallContext()
	f := new(ScFunc)
	f.ctx = ctx
	f.initView(hContract, hFunction)
	return f
}

func (f *ScFunc) Call() {
	if f.delay != 0 {
		Panic("cannot delay a call")
	}
	f.call(f.transfer)
}

func (f *ScFunc) Delay(seconds uint32) *ScFunc {
	f.delay = seconds
	return f
}

func (f *ScFunc) OfContract(hContract wasmtypes.ScHname) *ScFunc {
	f.hContract = hContract
	return f
}

func (f *ScFunc) Post() {
	f.PostToChain(f.ctx.ChainID())
}

func (f *ScFunc) PostToChain(chainID wasmtypes.ScChainID) {
	req := wasmrequests.PostRequest{
		ChainID:  chainID,
		Contract: f.hContract,
		Function: f.hFunction,
		Params:   f.params.Bytes(),
		Transfer: f.transfer.Bytes(),
		Delay:    f.delay,
	}
	res := Sandbox(FnPost, req.Bytes())
	if f.resultsProxy != nil {
		*f.resultsProxy = wasmtypes.NewProxy(NewScDictFromBytes(res))
	}
}

func (f *ScFunc) Transfer(transfer ScTransfers) *ScFunc {
	f.transfer = ScAssets(transfer)
	return f
}

func (f *ScFunc) TransferIotas(amount uint64) *ScFunc {
	f.transfer = make(ScAssets)
	f.transfer[wasmtypes.IOTA] = amount
	return f
}
