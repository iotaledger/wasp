// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// package coreutil provides functions to describe interface of the core contract
// in a compact way
package coreutil

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
)

// ContractInterface represents smart contract interface
type ContractInterface struct {
	Name        string
	Description string
	ProgramHash hashing.HashValue
}

type ContractProcessor struct {
	Interface *ContractInterface
	Handlers  map[iscp.Hname]ContractFunctionHandler
}

// ContractFunctionInterface represents entry point interface
type ContractFunctionInterface struct {
	Name   string
	IsView bool
}

// ContractFunctionHandler is a union structure: one of Handler, ViewHandler will be set
type ContractFunctionHandler struct {
	Interface   *ContractFunctionInterface
	Handler     Handler
	ViewHandler ViewHandler
}

type (
	Handler     func(ctx iscp.Sandbox) (dict.Dict, error)
	ViewHandler func(ctx iscp.SandboxView) (dict.Dict, error)
)

var (
	FuncFallback           = Func("fallbackHandler")
	FuncDefaultInitializer = Func("initializer")
)

func defaultInitFunc(ctx iscp.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("default init function invoked for contract %s from caller %s", ctx.Contract(), ctx.Caller())
	return nil, nil
}

func fallbackHandler(ctx iscp.Sandbox) (dict.Dict, error) {
	transferStr := "(empty)"
	if ctx.IncomingTransfer() != nil {
		transferStr = ctx.IncomingTransfer().String()
	}
	ctx.Log().Debugf("default full entry point handler invoked for contact %s from caller %s\nTransfer: %s",
		ctx.Contract(), ctx.Caller(), transferStr)
	return nil, nil
}

func NewContractInterface(name, description string) *ContractInterface {
	return &ContractInterface{
		Name:        name,
		Description: description,
		ProgramHash: hashing.HashStrings(name),
	}
}

// Func declares a full entry point
func Func(name string) ContractFunctionInterface {
	return ContractFunctionInterface{
		Name:   name,
		IsView: false,
	}
}

// Func declares a view entry point
func ViewFunc(name string) ContractFunctionInterface {
	return ContractFunctionInterface{
		Name:   name,
		IsView: true,
	}
}

// HandlerFunc declares a full entry point
func (fi *ContractFunctionInterface) Handler(fn Handler) ContractFunctionHandler {
	if fi.IsView {
		panic("can't create a full entry point handler from a view entry point")
	}
	return ContractFunctionHandler{Interface: fi, Handler: fn}
}

// ViewHandlerFunc declares a view entry point
func (fi *ContractFunctionInterface) ViewHandler(fn ViewHandler) ContractFunctionHandler {
	if !fi.IsView {
		panic("can't create a view entry point handler from a full entry point")
	}
	return ContractFunctionHandler{Interface: fi, ViewHandler: fn}
}

// Processor creates a ContractProcessor with the provided handlers
func (i *ContractInterface) Processor(init Handler, fns ...ContractFunctionHandler) *ContractProcessor {
	if init == nil {
		init = defaultInitFunc
	}
	handlers := map[iscp.Hname]ContractFunctionHandler{
		// under hname == 0 always resides default handler:
		0: FuncFallback.Handler(fallbackHandler),
		// constructor:
		iscp.EntryPointInit: FuncDefaultInitializer.Handler(init),
	}
	for _, f := range fns {
		hname := f.Interface.Hname()
		if _, ok := handlers[hname]; ok {
			panic(fmt.Sprintf("Duplicate function: %s (%s)", f.Interface.Name, hname.String()))
		}

		n := 0
		if f.Handler != nil {
			n++
		}
		if f.ViewHandler != nil {
			n++
		}
		if n != 1 {
			panic("Exactly one of (Handler, ViewHandler) must be set")
		}

		handlers[hname] = f
	}
	return &ContractProcessor{Interface: i, Handlers: handlers}
}

func (i *ContractProcessor) GetEntryPoint(code iscp.Hname) (iscp.VMProcessorEntryPoint, bool) {
	f, ok := i.Handlers[code]
	if !ok {
		return nil, false
	}
	return &f, true
}

func (i *ContractProcessor) GetDefaultEntryPoint() iscp.VMProcessorEntryPoint {
	ret := i.Handlers[0]
	return &ret
}

func (i *ContractProcessor) GetDescription() string {
	return i.Interface.Description
}

func (i *ContractInterface) Hname() iscp.Hname {
	return CoreHname(i.Name)
}

func (fi *ContractFunctionInterface) Hname() iscp.Hname {
	return iscp.Hn(fi.Name)
}

func (f *ContractFunctionHandler) Call(ctx interface{}) (dict.Dict, error) {
	switch tctx := ctx.(type) {
	case iscp.Sandbox:
		if f.Handler != nil {
			return f.Handler(tctx)
		}
	case iscp.SandboxView:
		if f.ViewHandler != nil {
			return f.ViewHandler(tctx)
		}
	}
	panic("inconsistency: wrong type of call context")
}

func (f *ContractFunctionHandler) IsView() bool {
	return f.ViewHandler != nil
}

func (i *ContractProcessor) GetStateReadOnly(chainState kv.KVStoreReader) kv.KVStoreReader {
	return subrealm.NewReadOnly(chainState, kv.Key(i.Interface.Hname().Bytes()))
}
