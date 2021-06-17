// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import (
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/chainid"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
)

var typeIds = map[int32]int32{
	wasmhost.KeyAccountId:       wasmhost.OBJTYPE_AGENT_ID,
	wasmhost.KeyBalances:        wasmhost.OBJTYPE_MAP,
	wasmhost.KeyCall:            wasmhost.OBJTYPE_BYTES,
	wasmhost.KeyCaller:          wasmhost.OBJTYPE_AGENT_ID,
	wasmhost.KeyChainId:         wasmhost.OBJTYPE_CHAIN_ID,
	wasmhost.KeyChainOwnerId:    wasmhost.OBJTYPE_AGENT_ID,
	wasmhost.KeyContract:        wasmhost.OBJTYPE_HNAME,
	wasmhost.KeyContractCreator: wasmhost.OBJTYPE_AGENT_ID,
	wasmhost.KeyDeploy:          wasmhost.OBJTYPE_BYTES,
	wasmhost.KeyEvent:           wasmhost.OBJTYPE_STRING,
	wasmhost.KeyExports:         wasmhost.OBJTYPE_STRING | wasmhost.OBJTYPE_ARRAY,
	wasmhost.KeyIncoming:        wasmhost.OBJTYPE_MAP,
	wasmhost.KeyLog:             wasmhost.OBJTYPE_STRING,
	wasmhost.KeyMaps:            wasmhost.OBJTYPE_MAP | wasmhost.OBJTYPE_ARRAY,
	wasmhost.KeyMinted:          wasmhost.OBJTYPE_MAP,
	wasmhost.KeyPanic:           wasmhost.OBJTYPE_STRING,
	wasmhost.KeyParams:          wasmhost.OBJTYPE_MAP,
	wasmhost.KeyPost:            wasmhost.OBJTYPE_BYTES,
	wasmhost.KeyRequestId:       wasmhost.OBJTYPE_REQUEST_ID,
	wasmhost.KeyResults:         wasmhost.OBJTYPE_MAP,
	wasmhost.KeyReturn:          wasmhost.OBJTYPE_MAP,
	wasmhost.KeyState:           wasmhost.OBJTYPE_MAP,
	wasmhost.KeyTimestamp:       wasmhost.OBJTYPE_INT64,
	wasmhost.KeyTrace:           wasmhost.OBJTYPE_STRING,
	wasmhost.KeyTransfers:       wasmhost.OBJTYPE_MAP | wasmhost.OBJTYPE_ARRAY,
	wasmhost.KeyUtility:         wasmhost.OBJTYPE_MAP,
}

type ScContext struct {
	ScSandboxObject
}

func NewScContext(vm *WasmProcessor) *ScContext {
	o := &ScContext{}
	o.vm = vm
	o.host = &vm.KvStoreHost
	o.name = "root"
	o.id = 1
	o.isRoot = true
	o.objects = make(map[int32]int32)
	return o
}

func (o *ScContext) Exists(keyID, typeID int32) bool {
	if keyID == wasmhost.KeyExports {
		return o.vm.ctx == nil && o.vm.ctxView == nil
	}
	return o.GetTypeID(keyID) > 0
}

func (o *ScContext) GetBytes(keyID, typeID int32) []byte {
	switch keyID {
	case wasmhost.KeyAccountId:
		return o.vm.accountID().Bytes()
	case wasmhost.KeyCaller:
		return o.vm.ctx.Caller().Bytes()
	case wasmhost.KeyChainId:
		return o.vm.chainID().Bytes()
	case wasmhost.KeyChainOwnerId:
		return o.vm.chainOwnerID().Bytes()
	case wasmhost.KeyContract:
		return o.vm.contract().Bytes()
	case wasmhost.KeyContractCreator:
		return o.vm.contractCreator().Bytes()
	case wasmhost.KeyRequestId:
		return o.vm.ctx.RequestID().Bytes()
	case wasmhost.KeyTimestamp:
		return codec.EncodeInt64(o.vm.ctx.GetTimestamp())
	}
	o.invalidKey(keyID)
	return nil
}

func (o *ScContext) GetObjectID(keyID, typeID int32) int32 {
	if keyID == wasmhost.KeyExports && (o.vm.ctx != nil || o.vm.ctxView != nil) {
		// once map has entries (after on_load) this cannot be called any more
		o.invalidKey(keyID)
		return 0
	}

	return GetMapObjectID(o, keyID, typeID, ObjFactories{
		wasmhost.KeyBalances:  func() WaspObject { return NewScBalances(o.vm, keyID) },
		wasmhost.KeyExports:   func() WaspObject { return NewScExports(o.vm) },
		wasmhost.KeyIncoming:  func() WaspObject { return NewScBalances(o.vm, keyID) },
		wasmhost.KeyMaps:      func() WaspObject { return NewScMaps(o.vm) },
		wasmhost.KeyMinted:    func() WaspObject { return NewScBalances(o.vm, keyID) },
		wasmhost.KeyParams:    func() WaspObject { return NewScDictFromKvStore(&o.vm.KvStoreHost, o.vm.params()) },
		wasmhost.KeyResults:   func() WaspObject { return NewScDict(o.vm) },
		wasmhost.KeyReturn:    func() WaspObject { return NewScDict(o.vm) },
		wasmhost.KeyState:     func() WaspObject { return NewScDictFromKvStore(&o.vm.KvStoreHost, o.vm.state()) },
		wasmhost.KeyTransfers: func() WaspObject { return NewScTransfers(o.vm) },
		wasmhost.KeyUtility:   func() WaspObject { return NewScUtility(o.vm) },
	})
}

func (o *ScContext) GetTypeID(keyID int32) int32 {
	return typeIds[keyID]
}

func (o *ScContext) SetBytes(keyID, typeID int32, bytes []byte) {
	switch keyID {
	case wasmhost.KeyCall:
		o.processCall(bytes)
	case wasmhost.KeyDeploy:
		o.processDeploy(bytes)
	case wasmhost.KeyEvent:
		o.vm.ctx.Event(string(bytes))
	case wasmhost.KeyLog:
		o.vm.log().Infof(string(bytes))
	case wasmhost.KeyTrace:
		o.vm.log().Debugf(string(bytes))
	case wasmhost.KeyPanic:
		o.vm.log().Panicf(string(bytes))
	case wasmhost.KeyPost:
		o.processPost(bytes)
	default:
		o.invalidKey(keyID)
	}
}

func (o *ScContext) processCall(bytes []byte) {
	decode := NewBytesDecoder(bytes)
	contract, err := coretypes.HnameFromBytes(decode.Bytes())
	if err != nil {
		o.Panic(err.Error())
	}
	function, err := coretypes.HnameFromBytes(decode.Bytes())
	if err != nil {
		o.Panic(err.Error())
	}
	params := o.getParams(int32(decode.Int64()))
	transfer := o.getTransfer(int32(decode.Int64()))

	o.Tracef("CALL c'%s' f'%s'", contract.String(), function.String())
	var results dict.Dict
	if o.vm.ctx != nil {
		results, err = o.vm.ctx.Call(contract, function, params, transfer)
	} else {
		results, err = o.vm.ctxView.Call(contract, function, params)
	}
	if err != nil {
		o.Panic("failed to invoke call: %v", err)
	}
	resultsID := o.GetObjectID(wasmhost.KeyReturn, wasmhost.OBJTYPE_MAP)
	o.host.FindObject(resultsID).(*ScDict).kvStore = results
}

func (o *ScContext) processDeploy(bytes []byte) {
	decode := NewBytesDecoder(bytes)
	programHash, err := hashing.HashValueFromBytes(decode.Bytes())
	if err != nil {
		o.Panic(err.Error())
	}
	name := string(decode.Bytes())
	description := string(decode.Bytes())
	params := o.getParams(int32(decode.Int64()))
	o.Tracef("DEPLOY c'%s' f'%s'", name, description)
	err = o.vm.ctx.DeployContract(programHash, name, description, params)
	if err != nil {
		o.Panic("failed to deploy: %v", err)
	}
}

// TODO refactor
func (o *ScContext) processPost(bytes []byte) {
	decode := NewBytesDecoder(bytes)
	chainID, err := chainid.ChainIDFromBytes(decode.Bytes())
	if err != nil {
		o.Panic(err.Error())
	}
	contract, err := coretypes.HnameFromBytes(decode.Bytes())
	if err != nil {
		o.Panic(err.Error())
	}
	function, err := coretypes.HnameFromBytes(decode.Bytes())
	if err != nil {
		o.Panic(err.Error())
	}
	o.Tracef("POST c'%s' f'%s'", contract.String(), function.String())
	params := o.getParams(int32(decode.Int64()))
	transfer := o.getTransfer(int32(decode.Int64()))
	metadata := &coretypes.SendMetadata{
		TargetContract: contract,
		EntryPoint:     function,
		Args:           params,
	}
	delay := decode.Int64()
	if delay == 0 {
		if !o.vm.ctx.Send(chainID.AsAddress(), transfer, metadata) {
			o.Panic("failed to send to %s", chainID.AsAddress().String())
		}
		return
	}

	if delay < -1 {
		o.Panic("invalid delay: %d", delay)
	}

	timeLock := time.Unix(0, o.vm.ctx.GetTimestamp())
	timeLock = timeLock.Add(time.Duration(delay) * time.Second)
	options := coretypes.SendOptions{
		TimeLock: uint32(timeLock.Unix()),
	}
	if !o.vm.ctx.Send(chainID.AsAddress(), transfer, metadata, options) {
		o.Panic("failed to send to %s", chainID.AsAddress().String())
	}
}

func (o *ScContext) getParams(paramsID int32) dict.Dict {
	if paramsID == 0 {
		return dict.New()
	}
	params := o.host.FindObject(paramsID).(*ScDict).kvStore.(dict.Dict)
	params.MustIterate("", func(key kv.Key, value []byte) bool {
		o.Tracef("  PARAM '%s'", key)
		return true
	})
	return params
}

func (o *ScContext) getTransfer(transferID int32) *ledgerstate.ColoredBalances {
	if transferID == 0 {
		return ledgerstate.NewColoredBalances(map[ledgerstate.Color]uint64{})
	}
	transfer := make(map[ledgerstate.Color]uint64)
	transferDict := o.host.FindObject(transferID).(*ScDict).kvStore
	transferDict.MustIterate("", func(key kv.Key, value []byte) bool {
		color, _, err := codec.DecodeColor([]byte(key))
		if err != nil {
			o.Panic(err.Error())
		}
		amount, _, err := codec.DecodeUint64(value)
		if err != nil {
			o.Panic(err.Error())
		}
		o.Tracef("  XFER %d '%s'", amount, color.String())
		transfer[color] = amount
		return true
	})
	return ledgerstate.NewColoredBalances(transfer)
}
