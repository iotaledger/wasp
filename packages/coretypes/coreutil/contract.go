// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// package coreutil provides functions to describe interface of the core contract
// in a compact way
package coreutil

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

const DefaultHandler = "defaultHandler"

// ContractInterface represents smart contract interface
type ContractInterface struct {
	Name        string
	Description string
	ProgramHash hashing.HashValue
	Functions   map[coretypes.Hname]ContractFunctionInterface
}

// ContractFunctionInterface represents entry point interface
type ContractFunctionInterface struct {
	Name        string
	Handler     Handler
	ViewHandler ViewHandler
}

type Handler func(ctx coretypes.Sandbox) (dict.Dict, error)
type ViewHandler func(ctx coretypes.SandboxView) (dict.Dict, error)

func defaultInitFunc(ctx coretypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("default init function invoked for contract %s from caller %s", ctx.Contract(), ctx.Caller())
	return nil, nil
}

func defaultHandlerFunc(ctx coretypes.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("default full entry point handler invoked for contact %s from caller %s\nTokens: %s",
		ctx.Contract(), ctx.Caller(), ctx.IncomingTransfer().String())
	return nil, nil
}

// Funcs declares init entry point and a list of full and view entry points
func Funcs(init Handler, fns []ContractFunctionInterface, defaultHandler ...Handler) map[coretypes.Hname]ContractFunctionInterface {
	if init == nil {
		init = defaultInitFunc
	}
	ret := map[coretypes.Hname]ContractFunctionInterface{
		coretypes.EntryPointInit: Func("init", init),
	}
	for _, f := range fns {
		hname := f.Hname()
		if _, ok := ret[hname]; ok {
			panic(fmt.Sprintf("Duplicate function: %s (%s)", f.Name, hname.String()))
		}

		handlers := 0
		if f.Handler != nil {
			handlers += 1
		}
		if f.ViewHandler != nil {
			handlers += 1
		}
		if handlers != 1 {
			panic("Exactly one of (Handler, ViewHandler) must be set")
		}

		ret[hname] = f
	}
	def := defaultHandlerFunc
	if len(defaultHandler) > 0 {
		def = defaultHandler[0]
	}
	// under hname == 0 always resides default handler
	ret[0] = ContractFunctionInterface{
		Name:    DefaultHandler,
		Handler: def,
	}
	return ret
}

// Func declares a full entry point: its name and its handler
func Func(name string, handler Handler) ContractFunctionInterface {
	return ContractFunctionInterface{
		Name:    name,
		Handler: handler,
	}
}

// Func declares a view entry point: its name and its handler
func ViewFunc(name string, handler ViewHandler) ContractFunctionInterface {
	return ContractFunctionInterface{
		Name:        name,
		ViewHandler: handler,
	}
}

func (i *ContractInterface) WithFunctions(init Handler, funcs []ContractFunctionInterface) {
	i.Functions = Funcs(init, funcs)
}

func (i *ContractInterface) GetFunction(name string) (*ContractFunctionInterface, bool) {
	f, ok := i.Functions[coretypes.Hn(name)]
	return &f, ok
}

func (i *ContractInterface) GetEntryPoint(code coretypes.Hname) (coretypes.VMProcessorEntryPoint, bool) {
	f, entryPointFound := i.Functions[code]
	if !entryPointFound {
		return nil, false
	}
	return &f, true
}

func (i *ContractInterface) GetDefaultEntryPoint() coretypes.VMProcessorEntryPoint {
	ret := i.Functions[0]
	return &ret
}

func (i *ContractInterface) GetDescription() string {
	return i.Description
}

// Hname caches the value
func (i *ContractInterface) Hname() coretypes.Hname {
	return CoreHname(i.Name)
}

func (f *ContractFunctionInterface) Hname() coretypes.Hname {
	return coretypes.Hn(f.Name)
}

func (f *ContractFunctionInterface) Call(ctx interface{}) (dict.Dict, error) {
	switch tctx := ctx.(type) {
	case coretypes.Sandbox:
		if f.Handler != nil {
			return f.Handler(tctx)
		}
	case coretypes.SandboxView:
		if f.ViewHandler != nil {
			return f.ViewHandler(tctx)
		}
	}
	panic("inconsistency: wrong type of call context")
}

func (f *ContractFunctionInterface) IsView() bool {
	return f.ViewHandler != nil
}
