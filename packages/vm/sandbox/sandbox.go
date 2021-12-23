// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package sandbox

import (
	"math/big"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/gas"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/vmcontext"
)

type sandbox struct {
	vmctx *vmcontext.VMContext
}

var _ iscp.Sandbox = &sandbox{}

func init() {
	vmcontext.NewSandbox = func(vmctx *vmcontext.VMContext) iscp.Sandbox {
		return &sandbox{vmctx: vmctx}
	}
}

func (s *sandbox) AccountID() *iscp.AgentID {
	s.Burn(gas.GetContractContext)
	return s.vmctx.AccountID()
}

func (s *sandbox) BalanceIotas() uint64 {
	s.Burn(gas.GetContractContext)
	return s.vmctx.GetIotaBalance(s.vmctx.AccountID())
}

func (s *sandbox) BalanceNativeToken(id *iotago.NativeTokenID) *big.Int {
	s.Burn(gas.GetContractContext)
	return s.vmctx.GetNativeTokenBalance(s.vmctx.AccountID(), id)
}

func (s *sandbox) Assets() *iscp.Assets {
	s.Burn(gas.GetContractContext)
	return s.vmctx.GetAssets(s.vmctx.AccountID())
}

// Call calls an entry point of contract, passes parameters and funds
func (s *sandbox) Call(target, entryPoint iscp.Hname, params dict.Dict, transfer *iscp.Assets) (dict.Dict, error) {
	s.Burn(gas.CallContract)
	return s.vmctx.Call(target, entryPoint, params, transfer)
}

func (s *sandbox) Caller() *iscp.AgentID {
	s.Burn(gas.GetCallerData)
	return s.vmctx.Caller()
}

func (s *sandbox) ChainID() *iscp.ChainID {
	s.Burn(gas.GetContractContext)
	return s.vmctx.ChainID()
}

func (s *sandbox) ChainOwnerID() *iscp.AgentID {
	s.Burn(gas.GetContractContext)
	return s.vmctx.ChainOwnerID()
}

func (s *sandbox) Contract() iscp.Hname {
	s.Burn(gas.GetContractContext)
	return s.vmctx.CurrentContractHname()
}

func (s *sandbox) ContractCreator() *iscp.AgentID {
	s.Burn(gas.GetContractContext)
	return s.vmctx.ContractCreator()
}

// DeployContract deploys contract by the binary hash
// and calls "init" endpoint (constructor) with provided parameters
func (s *sandbox) DeployContract(programHash hashing.HashValue, name, description string, initParams dict.Dict) error {
	// gas is deducted in the root contract
	return s.vmctx.DeployContract(programHash, name, description, initParams)
}

func (s *sandbox) Event(msg string) {
	s.Burn(gas.LogEvent(msg))
	s.Log().Infof("event::%s -> '%s'", s.vmctx.CurrentContractHname(), msg)
	s.vmctx.MustSaveEvent(s.vmctx.CurrentContractHname(), msg)
}

func (s *sandbox) GetEntropy() hashing.HashValue {
	s.Burn(gas.GetEntropy)
	return s.vmctx.Entropy()
}

func (s *sandbox) Timestamp() int64 {
	s.Burn(gas.GetContractContext)
	return s.vmctx.Timestamp()
}

func (s *sandbox) IncomingTransfer() *iscp.Assets {
	s.Burn(gas.GetIncomingTransfer)
	return s.vmctx.IncomingTransfer()
}

func (s *sandbox) Log() iscp.LogInterface {
	// TODO should Log be disabled for wasm contracts? not much of a point in exposing internal logging
	return s.vmctx
}

func (s *sandbox) Params() dict.Dict {
	s.Burn(gas.GetRequestData)
	return s.vmctx.Params()
}

func (s *sandbox) Request() iscp.Request {
	s.Burn(gas.GetRequestData)
	return s.vmctx.Request()
}

func (s *sandbox) Send(par iscp.RequestParameters) bool {
	s.Burn(gas.SendL1Request)
	return s.vmctx.Send(par)
}

func (s *sandbox) State() kv.KVStore {
	return s.vmctx.State()
	//return s.vmctx.State(s.Burn)
}

func (s *sandbox) Utils() iscp.Utils {
	return NewUtils(s.Gas())
}

func (s *sandbox) BlockContext(construct func(ctx iscp.Sandbox) interface{}, onClose func(interface{})) interface{} {
	// doesn't have a gas burn, only used for internal (native) contracts
	return s.vmctx.BlockContext(s, construct, onClose)
}

func (s *sandbox) StateAnchor() *iscp.StateAnchor {
	s.Burn(gas.GetStateAnchorInfo)
	return s.vmctx.StateAnchor()
}

func (s *sandbox) Gas() iscp.Gas {
	return s
}

func (s *sandbox) Burn(gas uint64) {
	s.vmctx.GasBurn(gas)
}

func (s *sandbox) Budget() uint64 {
	return s.vmctx.GasBudgetLeft()
}

func (s *sandbox) Foundries() iscp.Foundries {
	return s
}

func (s *sandbox) CreateNew(scheme iotago.TokenScheme, tag iotago.TokenTag, maxSupply *big.Int) uint32 {
	return s.vmctx.CreateNewFoundry(scheme, tag, maxSupply)
}

func (s *sandbox) Destroy(sn uint32) {
	s.vmctx.DestroyFoundry(sn)
}

func (s *sandbox) GetOutput(sn uint32) *iotago.FoundryOutput {
	return s.vmctx.GetOutput(sn)
}

func (s *sandbox) ModifySupply(sn uint32, delta *big.Int) {
	s.vmctx.ModifySupply(sn, delta)
}
