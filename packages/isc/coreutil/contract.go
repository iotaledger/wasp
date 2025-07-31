// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

// package coreutil provides functions to describe interface of the core contract
// in a compact way
package coreutil

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/kv/subrealm"
)

type Handler[S isc.SandboxBase] func(ctx S) isc.CallArguments

//********************************* *********************************\\

// ContractInfo holds basic information about a native smart contract
type ContractInfo struct {
	Name string
}

func NewContract(name string) *ContractInfo {
	return &ContractInfo{
		Name: name,
	}
}

func defaultInitFunc(ctx isc.Sandbox) isc.CallArguments {
	ctx.Log().Debugf("default init function invoked for contract %s from caller %s", ctx.Contract(), ctx.Caller())
	return nil
}

// Processor creates a ContractProcessor with the provided handlers
func (i *ContractInfo) Processor(init Handler[isc.Sandbox], eps ...isc.ProcessorEntryPoint) *ContractProcessor {
	if init != nil {
		panic("init function no longer supported")
	}
	handlers := map[isc.Hname]isc.ProcessorEntryPoint{}
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

func (i *ContractInfo) Func(name string) EntryPointInfo[isc.Sandbox] {
	return EntryPointInfo[isc.Sandbox]{
		Contract: i,
		Name:     name,
		isView:   false,
	}
}

// ViewFunc declares a view entry point
func (i *ContractInfo) ViewFunc(name string) EntryPointInfo[isc.SandboxView] {
	return EntryPointInfo[isc.SandboxView]{
		Contract: i,
		Name:     name,
		isView:   true,
	}
}

func (i *ContractInfo) StateSubrealm(chainState kv.KVStore) kv.KVStore {
	return isc.ContractStateSubrealm(chainState, i.Hname())
}

func (i *ContractInfo) StateSubrealmR(chainState kv.KVStoreReader) kv.KVStoreReader {
	return isc.ContractStateSubrealmR(chainState, i.Hname())
}

// EntryPointInfo holds basic information about an entry point
type EntryPointInfo[S isc.SandboxBase] struct {
	Contract *ContractInfo
	Name     string
	isView   bool
}

func (ep *EntryPointInfo[S]) IsView() bool {
	return ep.isView
}

func (ep *EntryPointInfo[S]) ContractInfo() *ContractInfo {
	return ep.Contract
}

func (ep *EntryPointInfo[S]) String() string {
	return ep.Name
}

func (ep *EntryPointInfo[S]) Hname() isc.Hname {
	return isc.Hn(ep.Name)
}

func (ep *EntryPointInfo[S]) Message(params isc.CallArguments) isc.Message {
	return isc.NewMessage(ep.Contract.Hname(), ep.Hname(), params)
}

// WithHandler specifies the handler function for the entry point
func (ep *EntryPointInfo[S]) WithHandler(fn Handler[S]) *EntryPointHandler[S] {
	return &EntryPointHandler[S]{Info: ep, Handler: fn}
}

type EntryPointHandler[S isc.SandboxBase] struct {
	Info    *EntryPointInfo[S]
	Handler Handler[S]
}

var (
	_ isc.ProcessorEntryPoint = &EntryPointHandler[isc.Sandbox]{}
	_ isc.ProcessorEntryPoint = &EntryPointHandler[isc.SandboxView]{}
)

func (h *EntryPointHandler[S]) Call(ctx isc.SandboxBase) isc.CallArguments {
	return h.Handler(ctx.(S))
}

func (h *EntryPointHandler[S]) IsView() bool {
	return h.Info.isView
}

func (h *EntryPointHandler[S]) Name() string {
	return h.Info.Name
}

func (h *EntryPointHandler[S]) Hname() isc.Hname {
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
