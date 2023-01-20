// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmlib

import (
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmrequests"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

type ScFuncCallContext interface {
	InitFuncCallContext()
}

type ScViewCallContext interface {
	InitViewCallContext(hContract wasmtypes.ScHname) wasmtypes.ScHname
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScView struct {
	hContract    wasmtypes.ScHname
	hFunction    wasmtypes.ScHname
	params       *ScDict
	resultsProxy *wasmtypes.Proxy
}

func NewScView(ctx ScViewCallContext, hContract, hFunction wasmtypes.ScHname) *ScView {
	// allow context to override default hContract
	hContract = ctx.InitViewCallContext(hContract)
	v := new(ScView)
	v.initView(hContract, hFunction)
	return v
}

func NewCallParamsProxy(v *ScView) wasmtypes.Proxy {
	v.params = NewScDict()
	return v.params.AsProxy()
}

func NewCallResultsProxy(v *ScView, resultsProxy *wasmtypes.Proxy) {
	v.resultsProxy = resultsProxy
}

func (v *ScView) Call() {
	v.callWithAllowance(nil)
}

func (v *ScView) callWithAllowance(allowance *ScTransfer) {
	req := wasmrequests.CallRequest{
		Contract:  v.hContract,
		Function:  v.hFunction,
		Params:    v.params.Bytes(),
		Allowance: allowance.Bytes(),
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
}

func NewScInitFunc(ctx ScFuncCallContext, hContract, hFunction wasmtypes.ScHname) *ScInitFunc {
	f := new(ScInitFunc)
	f.initView(hContract, hFunction)
	if ctx != nil {
		ctx.InitFuncCallContext()
	}
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
	for k, v := range f.params.dict {
		params = append(params, k)
		params = append(params, v)
	}
	return params
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScFunc struct {
	ScView
	allowance *ScTransfer
	delay     uint32
	transfer  *ScTransfer
}

func NewScFunc(ctx ScFuncCallContext, hContract, hFunction wasmtypes.ScHname) *ScFunc {
	ctx.InitFuncCallContext()
	f := new(ScFunc)
	f.initView(hContract, hFunction)
	return f
}

// Allowance defines the assets that the SC is allowed to take out of the caller account on the chain.
// Note that that does not mean that the SC will take them all. The SC needs to be able to take them explicitly.
// Otherwise the assets will stay in the caller’s account.
func (f *ScFunc) Allowance(allowance *ScTransfer) *ScFunc {
	f.allowance = allowance
	return f
}

func (f *ScFunc) AllowanceBaseTokens(amount uint64) *ScFunc {
	f.allowance = NewScTransferBaseTokens(amount)
	return f
}

func (f *ScFunc) Call() {
	if f.transfer != nil {
		Panic("cannot transfer assets in a call")
	}
	if f.delay != 0 {
		Panic("cannot delay a call")
	}
	f.callWithAllowance(f.allowance)
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
	f.PostToChain(ScFuncContext{}.CurrentChainID())
}

func (f *ScFunc) PostToChain(chainID wasmtypes.ScChainID) {
	req := wasmrequests.PostRequest{
		ChainID:   chainID,
		Contract:  f.hContract,
		Function:  f.hFunction,
		Params:    f.params.Bytes(),
		Allowance: f.allowance.Bytes(),
		Transfer:  f.transfer.Bytes(),
		Delay:     f.delay,
	}
	res := Sandbox(FnPost, req.Bytes())
	if f.resultsProxy != nil {
		*f.resultsProxy = wasmtypes.NewProxy(NewScDictFromBytes(res))
	}
}

// Transfer defines the assets that are transferred from the caller’s L1 address to his L2 account.
// The SC cannot touch these unless explicitly allowed.
// Transfer only comes into play with on-ledger requests. Off-ledger requests cannot do a transfer.
func (f *ScFunc) Transfer(transfer *ScTransfer) *ScFunc {
	f.transfer = transfer
	return f
}

func (f *ScFunc) TransferBaseTokens(amount uint64) *ScFunc {
	f.transfer = NewScTransferBaseTokens(amount)
	return f
}
