// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0
package sandbox

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/coretypes"
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
func (s *sandbox) Call(contractHname coretypes.Hname, entryPoint coretypes.Hname, params codec.ImmutableCodec, transfer coretypes.ColoredBalances) (codec.ImmutableCodec, error) {
	return s.vmctx.Call(contractHname, entryPoint, params, transfer)
}

// general
func (s *sandbox) ChainID() coretypes.ChainID {
	return s.vmctx.ChainID()
}

func (s *sandbox) ChainOwnerID() coretypes.AgentID {
	return s.vmctx.ChainOwnerID()
}

func (s *sandbox) State() codec.MutableMustCodec {
	return s.vmctx.State()
}

func (s *sandbox) RequestID() coretypes.RequestID {
	return *s.vmctx.Request().RequestID()
}

// call context

func (s *sandbox) Params() codec.ImmutableCodec {
	return s.vmctx.Params()
}

func (s *sandbox) Caller() coretypes.AgentID {
	return s.vmctx.Caller()
}

func (s *sandbox) MyContractID() coretypes.ContractID {
	return s.vmctx.CurrentContractID()
}

func (s *sandbox) MyAgentID() coretypes.AgentID {
	return coretypes.NewAgentIDFromContractID(s.vmctx.CurrentContractID())
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

func (s *sandbox) TransferToAddress(targetAddr address.Address, transfer coretypes.ColoredBalances) bool {
	return s.vmctx.TransferToAddress(targetAddr, transfer)
}

func (s *sandbox) TransferCrossChain(targetAgentID coretypes.AgentID, targetChainID coretypes.ChainID, transfer coretypes.ColoredBalances) bool {
	return s.vmctx.TransferCrossChain(targetAgentID, targetChainID, transfer)
}

func (s *sandbox) PostRequest(par vmtypes.NewRequestParams) bool {
	return s.vmctx.PostRequest(par)
}

func (s *sandbox) PostRequestToSelf(reqCode coretypes.Hname, args dict.Dict) bool {
	return s.vmctx.PostRequestToSelf(reqCode, args)
}

func (s *sandbox) PostRequestToSelfWithDelay(entryPoint coretypes.Hname, args dict.Dict, delaySec uint32) bool {
	return s.vmctx.PostRequestToSelfWithDelay(entryPoint, args, delaySec)
}

func (s *sandbox) Event(msg string) {
	s.vmctx.EventPublisher().Publish(msg)
}

func (s *sandbox) Eventf(format string, args ...interface{}) {
	s.vmctx.EventPublisher().Publishf(format, args...)
}
