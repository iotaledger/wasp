// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import "github.com/iotaledger/wasp/packages/vm/wasmhost"

type ScContract struct {
	ScDict
}

func (o *ScContract) Exists(keyId int32) bool {
	return o.GetTypeId(keyId) > 0
}

func (o *ScContract) GetBytes(keyId int32) []byte {
	switch keyId {
	case wasmhost.KeyChain:
		id := o.vm.ctx.ChainID()
		return id[:]
	case wasmhost.KeyChainOwner:
		id := o.vm.ctx.ChainOwnerID()
		return id[:]
	case wasmhost.KeyCreator:
		id := o.vm.ctx.ContractCreator()
		return id[:]
	case wasmhost.KeyId:
		id := o.vm.contractID()
		return id[:]
	}
	return o.ScDict.GetBytes(keyId)
}

func (o *ScContract) GetString(keyId int32) string {
	switch keyId {
	case wasmhost.KeyDescription:
		//TODO currently always returns "Wasm VM smart contract processor"
		// ask core contract for contract description instead?
		return o.vm.GetDescription()
	case wasmhost.KeyName: //TODO ask core contract for contract name?
	}
	return o.ScDict.GetString(keyId)
}

func (o *ScContract) GetTypeId(keyId int32) int32 {
	switch keyId {
	case wasmhost.KeyChain:
		return wasmhost.OBJTYPE_BYTES //TODO OBJTYPE_ADDRESS
	case wasmhost.KeyChainOwner:
		return wasmhost.OBJTYPE_BYTES //TODO OBJTYPE_AGENT
	case wasmhost.KeyCreator:
		return wasmhost.OBJTYPE_BYTES //TODO OBJTYPE_AGENT
	case wasmhost.KeyDescription:
		return wasmhost.OBJTYPE_STRING
	case wasmhost.KeyId:
		return wasmhost.OBJTYPE_BYTES
	case wasmhost.KeyName:
		return wasmhost.OBJTYPE_STRING
	}
	return 0
}
