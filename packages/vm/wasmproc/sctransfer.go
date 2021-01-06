// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/cbalances"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
)

type ScTransfer struct {
	ScDict
	agent coretypes.AgentID
	chain coretypes.ChainID
	balances map[balance.Color]int64
}

func (o *ScTransfer) Exists(keyId int32) bool {
	return false
}

func (o *ScTransfer) GetTypeId(keyId int32) int32 {
	return 0
}

func (o *ScTransfer) Invoke() {
    if o.chain != coretypes.NilChainID {
		o.vm.ctx.TransferCrossChain(o.agent, o.chain, cbalances.NewFromMap(o.balances))
		return
	}
	if o.agent.IsAddress() {
		o.vm.ctx.TransferToAddress(o.agent.MustAddress(), cbalances.NewFromMap(o.balances))
		return
	}
	for color,amount := range o.balances {
		o.vm.ctx.MoveTokens(o.agent, color, amount)
	}
}

func (o *ScTransfer) SetBytes(keyId int32, value []byte) {
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
		o.Panic("SetBytes: invalid access")
	}
}

func (o *ScTransfer) SetInt(keyId int32, value int64) {
	switch keyId {
	case wasmhost.KeyLength:
		o.agent = coretypes.AgentID{}
	default:
		key := o.vm.GetKeyFromId(keyId)
		color, _, err := balance.ColorFromBytes(key)
		if err != nil {
			o.Panic("SetInt: invalid color: " + err.Error())
		}
		if value > 0 {
			o.Trace("TRANSFER #%d c'%s' a'%s'", value, color.String(), o.agent.String())
            o.balances[color] = value
            return
		}
		if value == -1 && color == balance.ColorNew {
			o.Invoke()
			return
		}
		o.Panic("SetInt: invalid amount: " + err.Error())
	}
}

// \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\ // \\

type ScTransfers struct {
	ScDict
}

func (a *ScTransfers) GetObjectId(keyId int32, typeId int32) int32 {
	return GetArrayObjectId(a, keyId, typeId, func() WaspObject {
		return &ScTransfer{ balances: make(map[balance.Color]int64) }
	})
}

func (a *ScTransfers) SetString(keyId int32, value string) {
	a.Panic("SetString: Invalid access")
}
