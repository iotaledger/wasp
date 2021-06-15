// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package sandbox

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/chainid"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/sandbox/sandbox_utils"
	"github.com/iotaledger/wasp/packages/vm/vmcontext"
)

type sandboxView struct {
	vmctx *vmcontext.VMContext
}

var _ coretypes.SandboxView = sandboxView{}

func init() {
	vmcontext.NewSandboxView = func(vmctx *vmcontext.VMContext) coretypes.SandboxView {
		return sandboxView{vmctx}
	}
}

func (s sandboxView) AccountID() *coretypes.AgentID {
	return s.vmctx.AccountID()
}

func (s sandboxView) Balances() *ledgerstate.ColoredBalances {
	return s.vmctx.GetMyBalances()
}

func (s sandboxView) Call(contractHname coretypes.Hname, entryPoint coretypes.Hname, params dict.Dict) (dict.Dict, error) {
	return s.vmctx.Call(contractHname, entryPoint, params, nil)
}

func (s sandboxView) ChainID() *chainid.ChainID {
	return s.vmctx.ChainID()
}

func (s sandboxView) ChainOwnerID() *coretypes.AgentID {
	return s.vmctx.ChainOwnerID()
}

func (s sandboxView) Contract() coretypes.Hname {
	return s.vmctx.CurrentContractHname()
}

func (s sandboxView) ContractCreator() *coretypes.AgentID {
	return s.vmctx.ContractCreator()
}

func (s sandboxView) GetTimestamp() int64 {
	return s.vmctx.Timestamp()
}

func (s sandboxView) Log() coretypes.LogInterface {
	return s.vmctx
}

func (s sandboxView) Params() dict.Dict {
	return s.vmctx.Params()
}

func (s sandboxView) State() kv.KVStoreReader {
	return s.vmctx.State()
}

func (s sandboxView) Utils() coretypes.Utils {
	return sandbox_utils.NewUtils()
}
