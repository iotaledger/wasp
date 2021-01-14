// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
)

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScTransfers struct {
	ScSandboxObject
}

func NewScTransfers(vm *wasmProcessor) *ScTransfers {
	o := &ScTransfers{}
	o.vm = vm
	return o
}

func (a *ScTransfers) GetObjectId(keyId int32, typeId int32) int32 {
	return GetArrayObjectId(a, keyId, typeId, func() WaspObject {
		return NewScTransferInfo(a.vm)
	})
}

func (a *ScTransfers) SetString(keyId int32, value string) {
	a.Panic("SetString: Invalid access")
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScTransferInfo struct {
	ScSandboxObject
	agent    coretypes.AgentID
	balances map[balance.Color]int64
	chain    coretypes.ChainID
}

func NewScTransferInfo(vm *wasmProcessor) *ScTransferInfo {
	o := &ScTransferInfo{}
	o.vm = vm
	o.balances = make(map[balance.Color]int64)
	return o
}

func (o *ScTransferInfo) Exists(keyId int32) bool {
	return false
}

func (o *ScTransferInfo) GetTypeId(keyId int32) int32 {
	return 0
}

// TODO: by wrapping it all into one call is changing semantics of calls. It is not correct.
//  also because MoveTokens is synchronous, TransferCrossChain requires non empty
//  sender's balances and probably fees and TransferToAddress do not need it
//  The deprecated methods will be removed anyway. Only TransferToAddress will stay in the sandbox
func (o *ScTransferInfo) Invoke() {
	// is this a local transfer?
	if o.chain == o.vm.ctx.ChainID() {
		for color, amount := range o.balances {
			o.vm.ctx.MoveTokens(o.agent, color, amount)
		}
		return
	}
	// is this a cross chain transfer?
	if o.chain != coretypes.NilChainID {
		o.vm.ctx.TransferCrossChain(o.agent, o.chain, cbalances.NewFromMap(o.balances))
		return
	}
	// should be transfer to Tangle ledger address
	if o.agent.IsAddress() {
		o.vm.ctx.TransferToAddress(o.agent.MustAddress(), cbalances.NewFromMap(o.balances))
		return
	}
	o.Panic("Invoke: inconsistent agent id")
}

func (o *ScTransferInfo) SetBytes(keyId int32, value []byte) {
	var err error
	switch keyId {
	case wasmhost.KeyAgent:
		o.agent, err = coretypes.NewAgentIDFromBytes(value)
		if err != nil {
			o.Panic("SetBytes: invalid agent: " + err.Error())
		}
	case wasmhost.KeyChain:
		o.chain, err = coretypes.NewChainIDFromBytes(value)
		if err != nil {
			o.Panic("SetBytes: invalid chain: " + err.Error())
		}
	default:
		o.invalidKey(keyId)
	}
}

func (o *ScTransferInfo) SetInt(keyId int32, value int64) {
	switch keyId {
	case wasmhost.KeyLength:
		o.agent = coretypes.AgentID{}
	default:
		if value == 0 {
			// nothing to transfer
			return
		}
		key := o.host.GetKeyFromId(keyId)
		color, _, err := balance.ColorFromBytes(key)
		if err != nil {
			o.Panic("SetInt: invalid color: " + err.Error())
		}
		if value > 0 {
			o.Trace("TRANSFER #%d c'%s' a'%s'", value, color.String(), o.agent.String())
			o.balances[color] = value
			return
		}
		if value != -1 || color != balance.ColorNew {
			o.Panic("SetInt: invalid amount: " + err.Error())
		}
		o.Invoke()
	}
}
