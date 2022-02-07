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

// ContractInfo holds basic information about a native smart contract
type ContractInfo struct {
	Name        string
	Description string
	ProgramHash hashing.HashValue
}

func NewContract(name, description string) *ContractInfo {
	return &ContractInfo{
		Name:        name,
		Description: description,
		ProgramHash: hashing.HashStrings(name),
	}
}

// Processor creates a ContractProcessor with the provided handlers
func (i *ContractInfo) Processor(init Handler, eps ...ProcessorEntryPoint) *ContractProcessor {
	if init == nil {
		init = defaultInitFunc
	}
	handlers := map[iscp.Hname]ProcessorEntryPoint{
		// constructor:
		iscp.EntryPointInit: FuncDefaultInitializer.WithHandler(init),
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

func (i *ContractInfo) Hname() iscp.Hname {
	return CoreHname(i.Name)
}

type (
	ProcessorEntryPoint interface {
		iscp.VMProcessorEntryPoint
		Name() string
		Hname() iscp.Hname
	}

	Handler func(ctx iscp.Sandbox) dict.Dict

	// EntryPointInfo holds basic information about a full entry point
	EntryPointInfo struct{ Name string }

	EntryPointHandler struct {
		Info    *EntryPointInfo
		Handler Handler
	}

	ViewHandler func(ctx iscp.SandboxView) dict.Dict

	// ViewEntryPointInfo holds basic information about a view entry point
	ViewEntryPointInfo struct {
		Name string
	}

	ViewEntryPointHandler struct {
		Info    *ViewEntryPointInfo
		Handler ViewHandler
	}
)

var (
	_ ProcessorEntryPoint = &EntryPointHandler{}
	_ ProcessorEntryPoint = &ViewEntryPointHandler{}
)

// Func declares a full entry point
func Func(name string) EntryPointInfo {
	return EntryPointInfo{Name: name}
}

// WithHandler specifies the handler function for the entry point
func (ep *EntryPointInfo) WithHandler(fn Handler) *EntryPointHandler {
	return &EntryPointHandler{Info: ep, Handler: fn}
}

func (ep *EntryPointInfo) Hname() iscp.Hname {
	return iscp.Hn(ep.Name)
}

func (h *EntryPointHandler) Call(ctx interface{}) dict.Dict {
	return h.Handler(ctx.(iscp.Sandbox))
}

func (h *EntryPointHandler) IsView() bool {
	return false
}

func (h *EntryPointHandler) Name() string {
	return h.Info.Name
}

func (h *EntryPointHandler) Hname() iscp.Hname {
	return h.Info.Hname()
}

// ViewFunc declares a view entry point
func ViewFunc(name string) ViewEntryPointInfo {
	return ViewEntryPointInfo{Name: name}
}

// WithHandler specifies the handler function for the entry point
func (ep *ViewEntryPointInfo) WithHandler(fn ViewHandler) *ViewEntryPointHandler {
	return &ViewEntryPointHandler{Info: ep, Handler: fn}
}

func (ep *ViewEntryPointInfo) Hname() iscp.Hname {
	return iscp.Hn(ep.Name)
}

func (h *ViewEntryPointHandler) Call(ctx interface{}) dict.Dict {
	return h.Handler(ctx.(iscp.SandboxView))
}

func (h *ViewEntryPointHandler) IsView() bool {
	return true
}

func (h *ViewEntryPointHandler) Name() string {
	return h.Info.Name
}

func (h *ViewEntryPointHandler) Hname() iscp.Hname {
	return h.Info.Hname()
}

var (
	FuncFallback           = Func("fallbackHandler")
	FuncDefaultInitializer = Func("initializer")
)

func defaultInitFunc(ctx iscp.Sandbox) dict.Dict {
	ctx.Log().Debugf("default init function invoked for contract %s from caller %s", ctx.Contract(), ctx.Caller())
	return nil
}

type ContractProcessor struct {
	Contract *ContractInfo
	Handlers map[iscp.Hname]ProcessorEntryPoint
}

func (p *ContractProcessor) GetEntryPoint(code iscp.Hname) (iscp.VMProcessorEntryPoint, bool) {
	f, ok := p.Handlers[code]
	if !ok {
		return nil, false
	}
	return f, true
}

func (p *ContractProcessor) GetDescription() string {
	return p.Contract.Description
}

func (p *ContractProcessor) GetStateReadOnly(chainState kv.KVStoreReader) kv.KVStoreReader {
	return subrealm.NewReadOnly(chainState, kv.Key(p.Contract.Hname().Bytes()))
}
