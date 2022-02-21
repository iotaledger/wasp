// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package vmcontext

import (
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/packages/vm/sandbox"
)

type contractSandbox struct {
	sandbox.SandboxBase
}

func NewSandbox(vmctx *VMContext) iscp.Sandbox {
	ret := &contractSandbox{}
	ret.Ctx = vmctx
	return ret
}

// Call calls an entry point of contract, passes parameters and funds
func (s *contractSandbox) Call(target, entryPoint iscp.Hname, params dict.Dict, transfer *iscp.Assets) dict.Dict {
	s.Ctx.GasBurn(gas.BurnCodeCallContract)
	return s.Ctx.Call(target, entryPoint, params, transfer)
}

func (s *contractSandbox) Caller() *iscp.AgentID {
	s.Ctx.GasBurn(gas.BurnCodeGetCallerData)
	return s.Ctx.(*VMContext).Caller()
}

// DeployContract deploys contract by the binary hash
// and calls "init" endpoint (constructor) with provided parameters
func (s *contractSandbox) DeployContract(programHash hashing.HashValue, name, description string, initParams dict.Dict) {
	s.Ctx.(*VMContext).GasBurn(gas.BurnCodeDeployContract)
	s.Ctx.(*VMContext).DeployContract(programHash, name, description, initParams)
}

func (s *contractSandbox) Event(msg string) {
	s.Ctx.(*VMContext).GasBurn(gas.BurnCodeEmitEventFixed)
	s.Log().Infof("event::%s -> '%s'", s.Ctx.(*VMContext).CurrentContractHname(), msg)
	s.Ctx.(*VMContext).MustSaveEvent(s.Ctx.(*VMContext).CurrentContractHname(), msg)
}

func (s *contractSandbox) GetEntropy() hashing.HashValue {
	s.Ctx.(*VMContext).GasBurn(gas.BurnCodeGetContext)
	return s.Ctx.(*VMContext).Entropy()
}

func (s *contractSandbox) AllowanceAvailable() *iscp.Assets {
	s.Ctx.(*VMContext).GasBurn(gas.BurnCodeGetAllowance)
	return s.Ctx.(*VMContext).AllowanceAvailable()
}

func (s *contractSandbox) TransferAllowedFunds(target *iscp.AgentID, assets ...*iscp.Assets) *iscp.Assets {
	s.Ctx.(*VMContext).GasBurn(gas.BurnCodeTransferAllowance)
	return s.Ctx.(*VMContext).TransferAllowedFunds(target, false, assets...)
}

func (s *contractSandbox) TransferAllowedFundsForceCreateTarget(target *iscp.AgentID, assets ...*iscp.Assets) *iscp.Assets {
	s.Ctx.(*VMContext).GasBurn(gas.BurnCodeTransferAllowance)
	return s.Ctx.(*VMContext).TransferAllowedFunds(target, true, assets...)
}

func (s *contractSandbox) Request() iscp.Calldata {
	s.Ctx.(*VMContext).GasBurn(gas.BurnCodeGetContext)
	return s.Ctx.(*VMContext).Request()
}

func (s *contractSandbox) Send(par iscp.RequestParameters) {
	s.Ctx.(*VMContext).GasBurn(gas.BurnCodeSendL1Request, uint64(s.Ctx.(*VMContext).NumPostedOutputs))
	s.Ctx.(*VMContext).Send(par)
}

func (s *contractSandbox) EstimateRequiredDustDeposit(par iscp.RequestParameters) uint64 {
	// TODO missing gas burn
	return s.Ctx.(*VMContext).EstimateRequiredDustDeposit(par)
}

func (s *contractSandbox) State() kv.KVStore {
	return s.Ctx.(*VMContext).State()
}

func (s *contractSandbox) StateAnchor() *iscp.StateAnchor {
	s.Ctx.(*VMContext).GasBurn(gas.BurnCodeGetContext)
	return s.Ctx.(*VMContext).StateAnchor()
}

func (s *contractSandbox) RegisterError(messageFormat string) *iscp.VMErrorTemplate {
	return s.Ctx.(*VMContext).RegisterError(messageFormat)
}

// helper methods

func (s *contractSandbox) RequireCallerAnyOf(agentIDs []*iscp.AgentID, str ...string) {
	ok := false
	for _, agentID := range agentIDs {
		if s.Caller().Equals(agentID) {
			ok = true
		}
	}
	if !ok {
		if len(str) > 0 {
			s.Log().Panicf("'%s': unauthorized access", str[0])
		} else {
			s.Log().Panicf("unauthorized access")
		}
	}
}

func (s *contractSandbox) RequireCaller(agentID *iscp.AgentID, str ...string) {
	s.RequireCallerAnyOf([]*iscp.AgentID{agentID}, str...)
}

func (s *contractSandbox) RequireCallerIsChainOwner(str ...string) {
	s.RequireCaller(s.ChainOwnerID(), str...)
}

func (s *contractSandbox) Privileged() iscp.Privileged {
	return s
}

// privileged methods:

func (s *contractSandbox) TryLoadContract(programHash hashing.HashValue) error {
	return s.Ctx.(*VMContext).TryLoadContract(programHash)
}

func (s *contractSandbox) CreateNewFoundry(scheme iotago.TokenScheme, tag iotago.TokenTag, maxSupply *big.Int, metadata []byte) (uint32, uint64) {
	return s.Ctx.(*VMContext).CreateNewFoundry(scheme, tag, maxSupply, metadata)
}

func (s *contractSandbox) DestroyFoundry(sn uint32) uint64 {
	return s.Ctx.(*VMContext).DestroyFoundry(sn)
}

func (s *contractSandbox) ModifyFoundrySupply(sn uint32, delta *big.Int) int64 {
	return s.Ctx.(*VMContext).ModifyFoundrySupply(sn, delta)
}

func (s *contractSandbox) BlockContext(construct func(ctx iscp.Sandbox) interface{}, onClose func(interface{})) interface{} {
	// doesn't have a gas burn, only used for internal (native) contracts
	return s.Ctx.(*VMContext).BlockContext(s, construct, onClose)
}
