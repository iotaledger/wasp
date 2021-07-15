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

const DefaultHandler = "defaultHandler"

// ContractInterface represents smart contract interface
type ContractInterface struct {
	Name        string
	Description string
	ProgramHash hashing.HashValue
	Functions   map[iscp.Hname]ContractFunctionInterface
}

// ContractFunctionInterface represents entry point interface
type ContractFunctionInterface struct {
	Name        string
	Handler     Handler
	ViewHandler ViewHandler
}

type (
	Handler     func(ctx iscp.Sandbox) (dict.Dict, error)
	ViewHandler func(ctx iscp.SandboxView) (dict.Dict, error)
)

func defaultInitFunc(ctx iscp.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("default init function invoked for contract %s from caller %s", ctx.Contract(), ctx.Caller())
	return nil, nil
}

func defaultHandlerFunc(ctx iscp.Sandbox) (dict.Dict, error) {
	transferStr := "(empty)"
	if ctx.IncomingTransfer() != nil {
		transferStr = ctx.IncomingTransfer().String()
	}
	ctx.Log().Debugf("default full entry point handler invoked for contact %s from caller %s\nTransfer: %s",
		ctx.Contract(), ctx.Caller(), transferStr)
	return nil, nil
}

func NewContractInterface(name, description string, progHash ...hashing.HashValue) *ContractInterface {
	i := &ContractInterface{Name: name, Description: description}
	if len(progHash) > 0 {
		i.ProgramHash = progHash[0]
	} else {
		i.ProgramHash = hashing.HashStrings(i.Name)
	}
	return i
}

// Funcs declares init entry point and a list of full and view entry points
func Funcs(init Handler, fns []ContractFunctionInterface, defaultHandler ...Handler) map[iscp.Hname]ContractFunctionInterface {
	if init == nil {
		init = defaultInitFunc
	}
	ret := map[iscp.Hname]ContractFunctionInterface{
		iscp.EntryPointInit: Func("init", init),
	}
	for _, f := range fns {
		hname := f.Hname()
		if _, ok := ret[hname]; ok {
			panic(fmt.Sprintf("Duplicate function: %s (%s)", f.Name, hname.String()))
		}

		handlers := 0
		if f.Handler != nil {
			handlers++
		}
		if f.ViewHandler != nil {
			handlers++
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

func (i *ContractInterface) WithFunctions(init Handler, funcs []ContractFunctionInterface) *ContractInterface {
	i.Functions = Funcs(init, funcs)
	return i
}

func (i *ContractInterface) GetFunction(name string) (*ContractFunctionInterface, bool) {
	f, ok := i.Functions[iscp.Hn(name)]
	return &f, ok
}

func (i *ContractInterface) GetEntryPoint(code iscp.Hname) (iscp.VMProcessorEntryPoint, bool) {
	f, entryPointFound := i.Functions[code]
	if !entryPointFound {
		return nil, false
	}
	return &f, true
}

func (i *ContractInterface) GetDefaultEntryPoint() iscp.VMProcessorEntryPoint {
	ret := i.Functions[0]
	return &ret
}

func (i *ContractInterface) GetDescription() string {
	return i.Description
}

// Hname caches the value
func (i *ContractInterface) Hname() iscp.Hname {
	return CoreHname(i.Name)
}

func (f *ContractFunctionInterface) Hname() iscp.Hname {
	return iscp.Hn(f.Name)
}

func (f *ContractFunctionInterface) Call(ctx interface{}) (dict.Dict, error) {
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

func (f *ContractFunctionInterface) IsView() bool {
	return f.ViewHandler != nil
}

func (i *ContractInterface) GetStateReadOnly(chainState kv.KVStoreReader) kv.KVStoreReader {
	return subrealm.NewReadOnly(chainState, kv.Key(i.Hname().Bytes()))
}
