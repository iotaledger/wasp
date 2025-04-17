// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package vmimpl

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/eth/tracers"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/packages/vm/sandbox"
)

type contractSandbox struct {
	sandbox.SandboxBase
	reqctx *requestContext
}

func NewSandbox(reqctx *requestContext) isc.Sandbox {
	return &contractSandbox{
		SandboxBase: sandbox.SandboxBase{Ctx: reqctx},
		reqctx:      reqctx,
	}
}

// Call calls an entry point of contract, passes parameters and funds
func (s *contractSandbox) Call(msg isc.Message, allowance *isc.Assets) isc.CallArguments {
	s.Ctx.GasBurn(gas.BurnCodeCallContract) //nolint:typecheck
	return s.Ctx.Call(msg, allowance)
}

func (s *contractSandbox) Event(topic string, payload []byte) {
	s.Ctx.GasBurn(gas.BurnCodeEmitEvent1P, uint64(len(topic)+len(payload)))
	hContract := s.reqctx.CurrentContractHname()
	hex := hexutil.Encode(payload)
	if len(hex) > 80 {
		hex = hex[:40] + "..."
	}
	s.Log().Infof("event::%s -> %s(%s)", hContract.String(), topic, hex)
	s.reqctx.mustSaveEvent(hContract, topic, payload)
}

func (s *contractSandbox) GetEntropy() hashing.HashValue {
	s.Ctx.GasBurn(gas.BurnCodeGetContext)
	return s.reqctx.entropy
}

func (s *contractSandbox) AllowanceAvailable() *isc.Assets {
	s.Ctx.GasBurn(gas.BurnCodeGetAllowance)
	return s.reqctx.allowanceAvailable()
}

func (s *contractSandbox) TransferAllowedFunds(target isc.AgentID, transfer ...*isc.Assets) *isc.Assets {
	s.Ctx.GasBurn(gas.BurnCodeTransferAllowance)
	return s.reqctx.transferAllowedFunds(target, transfer...)
}

func (s *contractSandbox) Request() isc.Calldata {
	s.Ctx.GasBurn(gas.BurnCodeGetContext)
	return s.reqctx.req
}

func (s *contractSandbox) Send(par isc.RequestParameters) {
	s.Ctx.GasBurn(gas.BurnCodeSendL1Request, uint64(s.reqctx.numPostedOutputs))
	s.reqctx.send(par)
}

func (s *contractSandbox) State() kv.KVStore {
	return s.reqctx.contractStateWithGasBurn()
}

func (s *contractSandbox) StateAnchor() *isc.StateAnchor {
	s.Ctx.GasBurn(gas.BurnCodeGetContext)
	return s.reqctx.vm.stateAnchor()
}

func (s *contractSandbox) RequestIndex() uint16 {
	return s.reqctx.requestIndex
}

func (s *contractSandbox) RegisterError(messageFormat string) *isc.VMErrorTemplate {
	return s.reqctx.registerError(messageFormat)
}

func (s *contractSandbox) EVMTracer() *tracers.Tracer {
	return s.reqctx.vm.task.EVMTracer
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

func (s *contractSandbox) RequireCallerIsChainAdmin() {
	s.RequireCaller(s.ChainAdmin())
}

func (s *contractSandbox) Privileged() isc.Privileged {
	return s
}

// privileged methods:

func (s *contractSandbox) CreateNewFoundry(scheme isc.SimpleTokenScheme, metadata []byte) (uint32, uint64) {
	return s.reqctx.CreateNewFoundry(scheme, metadata)
}

func (s *contractSandbox) DestroyFoundry(sn uint32) uint64 {
	return s.reqctx.DestroyFoundry(sn)
}

func (s *contractSandbox) ModifyFoundrySupply(sn uint32, delta *big.Int) int64 {
	return s.reqctx.ModifyFoundrySupply(sn, delta)
}

func (s *contractSandbox) GasBurnEnable(enable bool) {
	s.Ctx.GasBurnEnable(enable)
}

func (s *contractSandbox) GasBurnEnabled() bool {
	return s.Ctx.GasBurnEnabled()
}

func (s *contractSandbox) MustMoveBetweenAccounts(fromAgentID, toAgentID isc.AgentID, assets *isc.Assets) {
	s.reqctx.mustMoveBetweenAccounts(fromAgentID, toAgentID, assets, true)
	s.checkRemainingTokens(fromAgentID)
}

func (s *contractSandbox) DebitFromAccount(agentID isc.AgentID, amount *big.Int) {
	s.reqctx.debitFromAccountFullDecimals(agentID, amount, true)
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

func (s *contractSandbox) CreditToAccount(agentID isc.AgentID, amount *big.Int) {
	s.reqctx.creditToAccountFullDecimals(agentID, amount, true)
}

func (s *contractSandbox) totalGasTokens() *isc.Assets {
	if s.reqctx.vm.task.EstimateGasMode {
		return isc.NewEmptyAssets()
	}
	amount := s.reqctx.gas.maxTokensToSpendForGasFee
	return isc.NewAssets(coin.Value(amount))
}

func (s *contractSandbox) CallOnBehalfOf(caller isc.AgentID, msg isc.Message, transfer *isc.Assets) isc.CallArguments {
	s.Ctx.GasBurn(gas.BurnCodeCallContract)
	return s.reqctx.CallOnBehalfOf(caller, msg, transfer)
}

func (s *contractSandbox) SendOnBehalfOf(caller isc.ContractIdentity, metadata isc.RequestParameters) {
	s.Ctx.GasBurn(gas.BurnCodeSendL1Request)
	s.reqctx.SendOnBehalfOf(caller, metadata)
}

func (s *contractSandbox) OnWriteReceipt(f isc.CoreCallbackFunc) {
	s.reqctx.onWriteReceipt = append(s.reqctx.onWriteReceipt, coreCallbackFunc{
		contract: s.reqctx.CurrentContractHname(),
		callback: f,
	})
}

func (s *contractSandbox) TakeStateSnapshot() int {
	s.reqctx.snapshots = append(s.reqctx.snapshots, stateSnapshot{
		txb:   s.reqctx.vm.createTxBuilderSnapshot(),
		state: s.reqctx.uncommittedState.Clone(),
	})
	return len(s.reqctx.snapshots) - 1
}

func (s *contractSandbox) RevertToStateSnapshot(i int) {
	if i < 0 || i >= len(s.reqctx.snapshots) {
		panic("invalid snapshot index")
	}
	s.reqctx.vm.restoreTxBuilderSnapshot(s.reqctx.snapshots[i].txb)
	s.reqctx.uncommittedState.SetMutations(s.reqctx.snapshots[i].state.Mutations())
}
