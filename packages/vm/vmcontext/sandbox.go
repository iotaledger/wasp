// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package vmcontext

import (
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/packages/vm/sandbox"
)

type contractSandbox struct {
	sandbox.SandboxBase
}

func NewSandbox(vmctx *VMContext) isc.Sandbox {
	ret := &contractSandbox{}
	ret.Ctx = vmctx
	return ret
}

// Call calls an entry point of contract, passes parameters and funds
func (s *contractSandbox) Call(target, entryPoint isc.Hname, params dict.Dict, transfer *isc.Assets) dict.Dict {
	s.Ctx.GasBurn(gas.BurnCodeCallContract)
	return s.Ctx.Call(target, entryPoint, params, transfer)
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

func (s *contractSandbox) AllowanceAvailable() *isc.Assets {
	s.Ctx.(*VMContext).GasBurn(gas.BurnCodeGetAllowance)
	return s.Ctx.(*VMContext).AllowanceAvailable()
}

func (s *contractSandbox) TransferAllowedFunds(target isc.AgentID, transfer ...*isc.Assets) *isc.Assets {
	s.Ctx.(*VMContext).GasBurn(gas.BurnCodeTransferAllowance)
	return s.Ctx.(*VMContext).TransferAllowedFunds(target, transfer...)
}

func (s *contractSandbox) TransferAllowedFundsForceCreateTarget(target isc.AgentID, transfer ...*isc.Assets) *isc.Assets {
	s.Ctx.(*VMContext).GasBurn(gas.BurnCodeTransferAllowance)
	return s.Ctx.(*VMContext).TransferAllowedFunds(target, transfer...)
}

func (s *contractSandbox) Request() isc.Calldata {
	s.Ctx.(*VMContext).GasBurn(gas.BurnCodeGetContext)
	return s.Ctx.(*VMContext).Request()
}

func (s *contractSandbox) Send(par isc.RequestParameters) {
	s.Ctx.(*VMContext).GasBurn(gas.BurnCodeSendL1Request, uint64(s.Ctx.(*VMContext).NumPostedOutputs))
	s.Ctx.(*VMContext).Send(par)
}

func (s *contractSandbox) SendAsNFT(par isc.RequestParameters, nftID iotago.NFTID) {
	s.Ctx.(*VMContext).GasBurn(gas.BurnCodeSendL1Request, uint64(s.Ctx.(*VMContext).NumPostedOutputs))
	s.Ctx.(*VMContext).SendAsNFT(par, nftID)
}

func (s *contractSandbox) EstimateRequiredStorageDeposit(par isc.RequestParameters) uint64 {
	s.Ctx.(*VMContext).GasBurn(gas.BurnCodeEstimateStorageDepositCost)
	return s.Ctx.(*VMContext).EstimateRequiredStorageDeposit(par)
}

func (s *contractSandbox) State() kv.KVStore {
	return s.Ctx.(*VMContext).State()
}

func (s *contractSandbox) StateAnchor() *isc.StateAnchor {
	s.Ctx.(*VMContext).GasBurn(gas.BurnCodeGetContext)
	return s.Ctx.(*VMContext).StateAnchor()
}

func (s *contractSandbox) RegisterError(messageFormat string) *isc.VMErrorTemplate {
	return s.Ctx.(*VMContext).RegisterError(messageFormat)
}

// helper methods

func (s *contractSandbox) RequireCallerAnyOf(agentIDs []isc.AgentID) {
	ok := false
	for _, agentID := range agentIDs {
		if s.Caller().Equals(agentID) {
			ok = true
		}
	}
	if !ok {
		panic(vm.ErrUnauthorized)
	}
}

func (s *contractSandbox) RequireCaller(agentID isc.AgentID) {
	if !s.Caller().Equals(agentID) {
		panic(vm.ErrUnauthorized)
	}
}

func (s *contractSandbox) RequireCallerIsChainOwner() {
	s.RequireCaller(s.ChainOwnerID())
}

func (s *contractSandbox) Privileged() isc.Privileged {
	return s
}

// privileged methods:

func (s *contractSandbox) TryLoadContract(programHash hashing.HashValue) error {
	return s.Ctx.(*VMContext).TryLoadContract(programHash)
}

func (s *contractSandbox) CreateNewFoundry(scheme iotago.TokenScheme, metadata []byte) (uint32, uint64) {
	return s.Ctx.(*VMContext).CreateNewFoundry(scheme, metadata)
}

func (s *contractSandbox) DestroyFoundry(sn uint32) uint64 {
	return s.Ctx.(*VMContext).DestroyFoundry(sn)
}

func (s *contractSandbox) ModifyFoundrySupply(sn uint32, delta *big.Int) int64 {
	return s.Ctx.(*VMContext).ModifyFoundrySupply(sn, delta)
}

func (s *contractSandbox) SubscribeBlockContext(openFunc, closeFunc isc.Hname) {
	s.Ctx.(*VMContext).SubscribeBlockContext(openFunc, closeFunc)
}

func (s *contractSandbox) SetBlockContext(bctx interface{}) {
	s.Ctx.(*VMContext).SetBlockContext(bctx)
}

func (s *contractSandbox) BlockContext() interface{} {
	return s.Ctx.(*VMContext).BlockContext()
}

func (s *contractSandbox) GasBurnEnable(enable bool) {
	s.Ctx.GasBurnEnable(enable)
}

func (s *contractSandbox) MustMoveBetweenAccounts(fromAgentID, toAgentID isc.AgentID, assets *isc.Assets) {
	s.Ctx.(*VMContext).mustMoveBetweenAccounts(fromAgentID, toAgentID, assets)
}

func (s *contractSandbox) DebitFromAccount(agentID isc.AgentID, tokens *isc.Assets) {
	s.Ctx.(*VMContext).debitFromAccount(agentID, tokens)
}

func (s *contractSandbox) CreditToAccount(agentID isc.AgentID, tokens *isc.Assets) {
	s.Ctx.(*VMContext).creditToAccount(agentID, tokens)
}

func (s *contractSandbox) TotalGasTokens() *isc.Assets {
	if s.Ctx.(*VMContext).task.EstimateGasMode {
		return isc.NewEmptyAssets()
	}
	amount := s.Ctx.(*VMContext).gasMaxTokensToSpendForGasFee
	nativeTokenID := s.Ctx.(*VMContext).chainInfo.GasFeePolicy.GasFeeTokenID
	if isc.IsEmptyNativeTokenID(nativeTokenID) {
		return isc.NewAssetsBaseTokens(amount)
	}
	return isc.NewAssets(0, iotago.NativeTokens{&iotago.NativeToken{
		ID:     nativeTokenID,
		Amount: new(big.Int).SetUint64(amount),
	}})
}
