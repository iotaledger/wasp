// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package sandbox

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/chainlog"
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
func (s *sandbox) CreateContract(programHash hashing.HashValue, name string, description string, initParams dict.Dict) error {
	return s.vmctx.CreateContract(programHash, name, description, initParams)
}

// Call calls an entry point of contact, passes parameters and funds
func (s *sandbox) Call(contractHname coretypes.Hname, entryPoint coretypes.Hname, params dict.Dict, transfer coretypes.ColoredBalances) (dict.Dict, error) {
	return s.vmctx.Call(contractHname, entryPoint, params, transfer)
}

// general
func (s *sandbox) ChainID() coretypes.ChainID {
	return s.vmctx.ChainID()
}

func (s *sandbox) ChainOwnerID() coretypes.AgentID {
	return s.vmctx.ChainOwnerID()
}

func (s *sandbox) ContractCreator() coretypes.AgentID {
	return s.vmctx.ContractCreator()
}

func (s *sandbox) State() kv.KVStore {
	return s.vmctx.State()
}

func (s *sandbox) RequestID() coretypes.RequestID {
	return s.vmctx.RequestID()
}

// call context

func (s *sandbox) Params() dict.Dict {
	return s.vmctx.Params()
}

func (s *sandbox) Caller() coretypes.AgentID {
	return s.vmctx.Caller()
}

func (s *sandbox) ContractID() coretypes.ContractID {
	return s.vmctx.CurrentContractID()
}

func (s *sandbox) AgentID() coretypes.AgentID {
	return coretypes.NewAgentIDFromContractID(s.vmctx.CurrentContractID())
}

func (s *sandbox) GetTimestamp() int64 {
	return s.vmctx.Timestamp()
}

func (s *sandbox) GetEntropy() hashing.HashValue {
	return s.vmctx.Entropy()
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

func (s *sandbox) Log() vmtypes.LogInterface {
	return s.vmctx
}

func (s *sandbox) Event(msg string) {
	s.vmctx.StoreToChainLog(s.vmctx.CurrentContractHname(), chainlog.TREvent, []byte(msg))
	s.vmctx.EventPublisher().Publish(msg)
}

func (s *sandbox) Eventf(format string, args ...interface{}) {
	s.vmctx.StoreToChainLog(s.vmctx.CurrentContractHname(), chainlog.TREvent, []byte(fmt.Sprintf(format, args...)))
	s.vmctx.EventPublisher().Publishf(format, args...)
}

func (s *sandbox) ChainLog(data []byte) {
	s.Log().Infof("chainlog record: '%s'", string(data))
	s.vmctx.StoreToChainLog(s.vmctx.CurrentContractHname(), chainlog.TRGenericData, data)
}
