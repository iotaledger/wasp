// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package sandbox

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/sandbox/sandbox_utils"
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
	return s.vmctx.AccountID()
}

func (s *sandbox) Balance(col colored.Color) uint64 {
	return s.vmctx.GetBalance(col)
}

func (s *sandbox) Balances() colored.Balances {
	return s.vmctx.GetMyBalances()
}

// Call calls an entry point of contract, passes parameters and funds
func (s *sandbox) Call(target, entryPoint iscp.Hname, params dict.Dict, transfer colored.Balances) (dict.Dict, error) {
	return s.vmctx.Call(target, entryPoint, params, transfer)
}

func (s *sandbox) Caller() *iscp.AgentID {
	return s.vmctx.Caller()
}

func (s *sandbox) ChainID() *iscp.ChainID {
	return s.vmctx.ChainID()
}

func (s *sandbox) ChainOwnerID() *iscp.AgentID {
	return s.vmctx.ChainOwnerID()
}

func (s *sandbox) Contract() iscp.Hname {
	return s.vmctx.CurrentContractHname()
}

func (s *sandbox) ContractCreator() *iscp.AgentID {
	return s.vmctx.ContractCreator()
}

// DeployContract deploys contract by the binary hash
// and calls "init" endpoint (constructor) with provided parameters
func (s *sandbox) DeployContract(programHash hashing.HashValue, name, description string, initParams dict.Dict) error {
	return s.vmctx.DeployContract(programHash, name, description, initParams)
}

func (s *sandbox) Event(msg string) {
	s.Log().Infof("event::%s -> '%s'", s.vmctx.CurrentContractHname(), msg)
	s.vmctx.MustSaveEvent(s.vmctx.CurrentContractHname(), msg)
}

func (s *sandbox) GetEntropy() hashing.HashValue {
	return s.vmctx.Entropy()
}

func (s *sandbox) GetTimestamp() int64 {
	return s.vmctx.Timestamp()
}

func (s *sandbox) IncomingTransfer() colored.Balances {
	return s.vmctx.GetIncoming()
}

func (s *sandbox) Log() iscp.LogInterface {
	return s.vmctx
}

func (s *sandbox) Minted() colored.Balances {
	return s.vmctx.Minted()
}

func (s *sandbox) Params() dict.Dict {
	return s.vmctx.Params()
}

func (s *sandbox) Request() iscp.Request {
	return s.vmctx.Request()
}

func (s *sandbox) Send(target ledgerstate.Address, tokens colored.Balances, metadata *iscp.SendMetadata, options ...iscp.SendOptions) bool {
	return s.vmctx.Send(target, tokens, metadata, options...)
}

func (s *sandbox) State() kv.KVStore {
	return s.vmctx.State()
}

func (s *sandbox) Utils() iscp.Utils {
	return sandbox_utils.NewUtils()
}

func (s *sandbox) BlockContext(construct func(ctx iscp.Sandbox) interface{}, onClose func(interface{})) interface{} {
	return s.vmctx.BlockContext(s, construct, onClose)
}

func (s *sandbox) StateAnchor() iscp.StateAnchor {
	return s.vmctx
}
