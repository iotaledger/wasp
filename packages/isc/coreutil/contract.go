// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// package coreutil provides functions to describe interface of the core contract
// in a compact way
package coreutil

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
)

type Handler func(ctx isc.Sandbox) dict.Dict

type ViewHandler func(ctx isc.SandboxView) dict.Dict

//********************************* *********************************\\

// ContractInfo holds basic information about a native smart contract
type ContractInfo struct {
	Name        string
	ProgramHash hashing.HashValue
}

var FuncDefaultInitializer = Func("initializer")

func NewContract(name string) *ContractInfo {
	return &ContractInfo{
		Name:        name,
		ProgramHash: CoreContractProgramHash(name),
	}
}

func CoreContractProgramHash(name string) hashing.HashValue {
	return hashing.HashStrings(name)
}

func defaultInitFunc(ctx isc.Sandbox) dict.Dict {
	ctx.Log().Debugf("default init function invoked for contract %s from caller %s", ctx.Contract(), ctx.Caller())
	return nil
}

// Processor creates a ContractProcessor with the provided handlers
func (i *ContractInfo) Processor(init Handler, eps ...isc.ProcessorEntryPoint) *ContractProcessor {
	if init == nil {
		init = defaultInitFunc
	}
	handlers := map[isc.Hname]isc.ProcessorEntryPoint{
		// constructor:
		isc.EntryPointInit: FuncDefaultInitializer.WithHandler(init),
	}
	for _, ep := range eps {
		hname := ep.Hname()
		if _, ok := handlers[hname]; ok {
			panic(fmt.Sprintf("Duplicate function: %s (%s)", ep.Name(), hname.String()))
		}

		handlers[hname] = ep
	}
	return &ContractProcessor{Contract: i, Handlers: handlers}
}

func (i *ContractInfo) Hname() isc.Hname {
	return CoreHname(i.Name)
}

// FullKey concatenates 4 bytes of hname with postfix
func (i *ContractInfo) FullKey(postfix []byte) []byte {
	return append(i.Hname().Bytes(), postfix...)
}

//********************************* *********************************\\

// EntryPointInfo holds basic information about a full entry point
type EntryPointInfo struct{ Name string }

// Func declares a full entry point
func Func(name string) EntryPointInfo {
	return EntryPointInfo{Name: name}
}

// WithHandler specifies the handler function for the entry point
func (ep *EntryPointInfo) WithHandler(fn Handler) *EntryPointHandler {
	return &EntryPointHandler{Info: ep, Handler: fn}
}

func (ep *EntryPointInfo) Hname() isc.Hname {
	return isc.Hn(ep.Name)
}

//********************************* *********************************\\

type EntryPointHandler struct {
	Info    *EntryPointInfo
	Handler Handler
}

var _ isc.ProcessorEntryPoint = &EntryPointHandler{}

func (h *EntryPointHandler) Call(ctx interface{}) dict.Dict {
	return h.Handler(ctx.(isc.Sandbox))
}

func (h *EntryPointHandler) IsView() bool {
	return false
}

func (h *EntryPointHandler) Name() string {
	return h.Info.Name
}

func (h *EntryPointHandler) Hname() isc.Hname {
	return h.Info.Hname()
}

//********************************* *********************************\\

// ViewEntryPointInfo holds basic information about a view entry point
type ViewEntryPointInfo struct {
	Name string
}

// ViewFunc declares a view entry point
func ViewFunc(name string) ViewEntryPointInfo {
	return ViewEntryPointInfo{Name: name}
}

// WithHandler specifies the handler function for the entry point
func (ep *ViewEntryPointInfo) WithHandler(fn ViewHandler) *ViewEntryPointHandler {
	return &ViewEntryPointHandler{Info: ep, Handler: fn}
}

func (ep *ViewEntryPointInfo) Hname() isc.Hname {
	return isc.Hn(ep.Name)
}

//********************************* *********************************\\

type ViewEntryPointHandler struct {
	Info    *ViewEntryPointInfo
	Handler ViewHandler
}

var _ isc.ProcessorEntryPoint = &ViewEntryPointHandler{}

func (h *ViewEntryPointHandler) Call(ctx interface{}) dict.Dict {
	return h.Handler(ctx.(isc.SandboxView))
}

func (h *ViewEntryPointHandler) IsView() bool {
	return true
}

func (h *ViewEntryPointHandler) Name() string {
	return h.Info.Name
}

func (h *ViewEntryPointHandler) Hname() isc.Hname {
	return h.Info.Hname()
}

//********************************* *********************************\\

type ContractProcessor struct {
	Contract *ContractInfo
	Handlers map[isc.Hname]isc.ProcessorEntryPoint
}

func (p *ContractProcessor) GetEntryPoint(code isc.Hname) (isc.VMProcessorEntryPoint, bool) {
	f, ok := p.Handlers[code]
	if !ok {
		return nil, false
	}
	return f, true
}

func (p *ContractProcessor) Entrypoints() map[isc.Hname]isc.ProcessorEntryPoint {
	return p.Handlers
}

func (p *ContractProcessor) GetStateReadOnly(chainState kv.KVStoreReader) kv.KVStoreReader {
	return subrealm.NewReadOnly(chainState, kv.Key(p.Contract.Hname().Bytes()))
}
