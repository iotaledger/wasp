// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package vmimpl

import (
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"

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

func NewSandbox(vmctx *vmContext) isc.Sandbox {
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
func (s *contractSandbox) DeployContract(programHash hashing.HashValue, name string, initParams dict.Dict) {
	s.Ctx.GasBurn(gas.BurnCodeDeployContract)
	s.Ctx.(*vmContext).DeployContract(programHash, name, initParams)
}

func (s *contractSandbox) Event(topic string, payload []byte) {
	s.Ctx.GasBurn(gas.BurnCodeEmitEventFixed)
	hContract := s.Ctx.(*vmContext).CurrentContractHname()
	hex := iotago.EncodeHex(payload)
	if len(hex) > 80 {
		hex = hex[:40] + "..."
	}
	s.Log().Infof("event::%s -> %s(%s)", hContract.String(), topic, hex)
	s.Ctx.(*vmContext).MustSaveEvent(hContract, topic, payload)
}

func (s *contractSandbox) GetEntropy() hashing.HashValue {
	s.Ctx.GasBurn(gas.BurnCodeGetContext)
	return s.Ctx.(*vmContext).Entropy()
}

func (s *contractSandbox) AllowanceAvailable() *isc.Assets {
	s.Ctx.GasBurn(gas.BurnCodeGetAllowance)
	return s.Ctx.(*vmContext).AllowanceAvailable()
}

func (s *contractSandbox) TransferAllowedFunds(target isc.AgentID, transfer ...*isc.Assets) *isc.Assets {
	s.Ctx.GasBurn(gas.BurnCodeTransferAllowance)
	return s.Ctx.(*vmContext).TransferAllowedFunds(target, transfer...)
}

func (s *contractSandbox) TransferAllowedFundsForceCreateTarget(target isc.AgentID, transfer ...*isc.Assets) *isc.Assets {
	s.Ctx.GasBurn(gas.BurnCodeTransferAllowance)
	return s.Ctx.(*vmContext).TransferAllowedFunds(target, transfer...)
}

func (s *contractSandbox) Request() isc.Calldata {
	s.Ctx.GasBurn(gas.BurnCodeGetContext)
	return s.Ctx.(*vmContext).Request()
}

func (s *contractSandbox) Send(par isc.RequestParameters) {
	s.Ctx.GasBurn(gas.BurnCodeSendL1Request, uint64(s.Ctx.(*vmContext).reqCtx.numPostedOutputs))
	s.Ctx.(*vmContext).Send(par)
}

func (s *contractSandbox) EstimateRequiredStorageDeposit(par isc.RequestParameters) uint64 {
	s.Ctx.GasBurn(gas.BurnCodeEstimateStorageDepositCost)
	return s.Ctx.(*vmContext).EstimateRequiredStorageDeposit(par)
}

func (s *contractSandbox) State() kv.KVStore {
	return s.Ctx.(*vmContext).State()
}

func (s *contractSandbox) StateAnchor() *isc.StateAnchor {
	s.Ctx.GasBurn(gas.BurnCodeGetContext)
	return s.Ctx.(*vmContext).StateAnchor()
}

func (s *contractSandbox) RegisterError(messageFormat string) *isc.VMErrorTemplate {
	return s.Ctx.(*vmContext).RegisterError(messageFormat)
}

func (s *contractSandbox) EVMTracer() *isc.EVMTracer {
	return s.Ctx.(*vmContext).task.EVMTracer
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
	return s.Ctx.(*vmContext).TryLoadContract(programHash)
}

func (s *contractSandbox) CreateNewFoundry(scheme iotago.TokenScheme, metadata []byte) (uint32, uint64) {
	return s.Ctx.(*vmContext).CreateNewFoundry(scheme, metadata)
}

func (s *contractSandbox) DestroyFoundry(sn uint32) uint64 {
	return s.Ctx.(*vmContext).DestroyFoundry(sn)
}

func (s *contractSandbox) ModifyFoundrySupply(sn uint32, delta *big.Int) int64 {
	return s.Ctx.(*vmContext).ModifyFoundrySupply(sn, delta)
}

func (s *contractSandbox) SetBlockContext(bctx interface{}) {
	s.Ctx.(*vmContext).SetBlockContext(bctx)
}

func (s *contractSandbox) BlockContext() interface{} {
	return s.Ctx.(*vmContext).BlockContext()
}

func (s *contractSandbox) GasBurnEnable(enable bool) {
	s.Ctx.GasBurnEnable(enable)
}

func (s *contractSandbox) MustMoveBetweenAccounts(fromAgentID, toAgentID isc.AgentID, assets *isc.Assets) {
	s.Ctx.(*vmContext).mustMoveBetweenAccounts(fromAgentID, toAgentID, assets)
	s.checkRemainingTokens(fromAgentID)
}

func (s *contractSandbox) DebitFromAccount(agentID isc.AgentID, tokens *isc.Assets) {
	s.Ctx.(*vmContext).debitFromAccount(agentID, tokens)
	s.checkRemainingTokens(agentID)
}

func (s *contractSandbox) checkRemainingTokens(debitedAccount isc.AgentID) {
	// assert that remaining tokens in the sender's account are enough to pay for the gas budget
	if debitedAccount.Equals(s.Request().SenderAccount()) && !s.HasInAccount(
		debitedAccount,
		s.totalGasTokens(),
	) {
		panic(vm.ErrNotEnoughTokensLeftForGas)
	}
}

func (s *contractSandbox) CreditToAccount(agentID isc.AgentID, tokens *isc.Assets) {
	s.Ctx.(*vmContext).creditToAccount(agentID, tokens)
}

func (s *contractSandbox) RetryUnprocessable(req isc.Request, blockIndex uint32, outputIndex uint16) {
	s.Ctx.(*vmContext).RetryUnprocessable(req, blockIndex, outputIndex)
}

func (s *contractSandbox) totalGasTokens() *isc.Assets {
	if s.Ctx.(*vmContext).task.EstimateGasMode {
		return isc.NewEmptyAssets()
	}
	amount := s.Ctx.(*vmContext).reqCtx.gas.maxTokensToSpendForGasFee
	return isc.NewAssetsBaseTokens(amount)
}

func (s *contractSandbox) CallOnBehalfOf(caller isc.AgentID, target, entryPoint isc.Hname, params dict.Dict, transfer *isc.Assets) dict.Dict {
	s.Ctx.GasBurn(gas.BurnCodeCallContract)
	return s.Ctx.(*vmContext).CallOnBehalfOf(caller, target, entryPoint, params, transfer)
}

func (s *contractSandbox) SetEVMFailed(tx *types.Transaction, receipt *types.Receipt) {
	s.Ctx.(*vmContext).SetEVMFailed(tx, receipt)
}