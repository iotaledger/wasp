// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0
package sandbox

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/coret"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/vmcontext"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

type sandbox struct {
	vmctx *vmcontext.VMContext
}

func init() {
	vmcontext.NewSandbox = new
}

func new(vmctx *vmcontext.VMContext) vmtypes.Sandbox {
	return &sandbox{
		vmctx: vmctx,
	}
}

// CreateContract deploys contract by the binary hash
// and calls "init" endpoint (constructor) with provided parameters
func (s *sandbox) CreateContract(programHash hashing.HashValue, name string, description string, initParams codec.ImmutableCodec) error {
	return s.vmctx.CreateContract(programHash, name, description, initParams)
}

// Call calls an entry point of contact, passes parameters and funds
func (s *sandbox) Call(contractHname coret.Hname, entryPoint coret.Hname, params codec.ImmutableCodec, transfer coret.ColoredBalances) (codec.ImmutableCodec, error) {
	return s.vmctx.Call(contractHname, entryPoint, params, transfer)
}

// general
func (s *sandbox) ChainID() coret.ChainID {
	return s.vmctx.ChainID()
}

func (s *sandbox) ChainOwnerID() coret.AgentID {
	return s.vmctx.ChainOwnerID()
}

func (s *sandbox) State() codec.MutableMustCodec {
	return s.vmctx.State()
}

func (s *sandbox) RequestID() coret.RequestID {
	return *s.vmctx.Request().RequestID()
}

// call context

func (s *sandbox) Params() codec.ImmutableCodec {
	return s.vmctx.Params()
}

func (s *sandbox) Caller() coret.AgentID {
	return s.vmctx.Caller()
}

func (s *sandbox) MyContractID() coret.ContractID {
	return s.vmctx.CurrentContractID()
}

func (s *sandbox) MyAgentID() coret.AgentID {
	return coret.NewAgentIDFromContractID(s.vmctx.CurrentContractID())
}

func (s *sandbox) GetTimestamp() int64 {
	return s.vmctx.Timestamp()
}

func (s *sandbox) GetEntropy() hashing.HashValue {
	return s.vmctx.Entropy()
}

func (s *sandbox) Panic(v interface{}) {
	panic(v)
}

func (s *sandbox) Rollback() {
	s.vmctx.Rollback()
}

func (s *sandbox) TransferToAddress(targetAddr address.Address, transfer coret.ColoredBalances) bool {
	return s.vmctx.TransferToAddress(targetAddr, transfer)
}

func (s *sandbox) TransferCrossChain(targetAgentID coret.AgentID, targetChainID coret.ChainID, transfer coret.ColoredBalances) bool {
	return s.vmctx.TransferCrossChain(targetAgentID, targetChainID, transfer)
}

func (s *sandbox) PostRequest(par vmtypes.NewRequestParams) bool {
	return s.vmctx.PostRequest(par)
}

func (s *sandbox) PostRequestToSelf(reqCode coret.Hname, args dict.Dict) bool {
	return s.vmctx.PostRequestToSelf(reqCode, args)
}

func (s *sandbox) PostRequestToSelfWithDelay(entryPoint coret.Hname, args dict.Dict, delaySec uint32) bool {
	return s.vmctx.PostRequestToSelfWithDelay(entryPoint, args, delaySec)
}

func (s *sandbox) Event(msg string) {
	s.vmctx.EventPublisher().Publish(msg)
}

func (s *sandbox) Eventf(format string, args ...interface{}) {
	s.vmctx.EventPublisher().Publishf(format, args...)
}
