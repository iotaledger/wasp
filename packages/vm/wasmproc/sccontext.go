// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import (
	"time"

	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/wasmhost"
)

var typeIds = map[int32]int32{
	wasmhost.KeyAccountID:       wasmhost.OBJTYPE_AGENT_ID,
	wasmhost.KeyBalances:        wasmhost.OBJTYPE_MAP,
	wasmhost.KeyCall:            wasmhost.OBJTYPE_BYTES,
	wasmhost.KeyCaller:          wasmhost.OBJTYPE_AGENT_ID,
	wasmhost.KeyChainID:         wasmhost.OBJTYPE_CHAIN_ID,
	wasmhost.KeyChainOwnerID:    wasmhost.OBJTYPE_AGENT_ID,
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
	wasmhost.KeyRandom:          wasmhost.OBJTYPE_BYTES,
	wasmhost.KeyRequestID:       wasmhost.OBJTYPE_REQUEST_ID,
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
	wc *WasmContext
}

func NewScContext(wc *WasmContext, host *wasmhost.KvStoreHost) *ScContext {
	o := &ScContext{}
	o.wc = wc
	o.host = host
	o.name = "root"
	o.id = 1
	o.isRoot = true
	o.objects = make(map[int32]int32)
	return o
}

func (o *ScContext) Exists(keyID, typeID int32) bool {
	if keyID == wasmhost.KeyExports {
		return o.wc.ctx == nil && o.wc.ctxView == nil
	}
	return o.GetTypeID(keyID) > 0
}

func (o *ScContext) GetBytes(keyID, typeID int32) []byte {
	if o.wc == nil {
		o.Panicf("missing context")
	}
	switch keyID {
	// common functionality
	case wasmhost.KeyAccountID:
		return o.wc.common.AccountID().Bytes()
	case wasmhost.KeyChainID:
		return o.wc.common.ChainID().Bytes()
	case wasmhost.KeyChainOwnerID:
		return o.wc.common.ChainOwnerID().Bytes()
	case wasmhost.KeyContract:
		return o.wc.common.Contract().Bytes()
	case wasmhost.KeyContractCreator:
		return o.wc.common.ContractCreator().Bytes()
	case wasmhost.KeyTimestamp:
		return codec.EncodeInt64(o.wc.common.GetTimestamp())
		// ctx-only functionality
	case wasmhost.KeyCaller:
		return o.wc.ctx.Caller().Bytes()
	case wasmhost.KeyRandom:
		return o.wc.ctx.GetEntropy().Bytes()
	case wasmhost.KeyRequestID:
		return o.wc.ctx.Request().ID().Bytes()
	}
	o.InvalidKey(keyID)
	return nil
}

func (o *ScContext) GetObjectID(keyID, typeID int32) int32 {
	if keyID == wasmhost.KeyExports && (o.wc.ctx != nil || o.wc.ctxView != nil) {
		// once map has entries (after on_load) this cannot be called any longer
		o.InvalidKey(keyID)
		return 0
	}

	return GetMapObjectID(o, keyID, typeID, ObjFactories{
		wasmhost.KeyBalances:  func() WaspObject { return NewScBalances(o.wc, keyID) },
		wasmhost.KeyExports:   func() WaspObject { return NewScExports(o.wc) },
		wasmhost.KeyIncoming:  func() WaspObject { return NewScBalances(o.wc, keyID) },
		wasmhost.KeyMaps:      func() WaspObject { return NewScMaps(o.host) },
		wasmhost.KeyMinted:    func() WaspObject { return NewScBalances(o.wc, keyID) },
		wasmhost.KeyParams:    func() WaspObject { return NewScDict(o.host, o.wc.params()) },
		wasmhost.KeyResults:   func() WaspObject { return NewScDict(o.host, dict.New()) },
		wasmhost.KeyReturn:    func() WaspObject { return NewScDict(o.host, dict.New()) },
		wasmhost.KeyState:     func() WaspObject { return NewScDict(o.host, o.wc.state()) },
		wasmhost.KeyTransfers: func() WaspObject { return NewScTransfers(o.wc) },
		wasmhost.KeyUtility:   func() WaspObject { return NewScUtility(o.wc, nil) },
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
		o.wc.ctx.Event(string(bytes))
	case wasmhost.KeyLog:
		o.wc.log().Infof(string(bytes))
	case wasmhost.KeyTrace:
		o.wc.log().Debugf(string(bytes))
	case wasmhost.KeyPanic:
		o.wc.log().Panicf(string(bytes))
	case wasmhost.KeyPost:
		o.processPost(bytes)
	default:
		o.InvalidKey(keyID)
	}
}

func (o *ScContext) processCall(bytes []byte) {
	decode := wasmlib.NewBytesDecoder(bytes)
	contract, err := iscp.HnameFromBytes(decode.Hname().Bytes())
	if err != nil {
		o.Panicf(err.Error())
	}
	function, err := iscp.HnameFromBytes(decode.Hname().Bytes())
	if err != nil {
		o.Panicf(err.Error())
	}
	params := o.getParams(decode.Int32())
	transfer := o.getTransfer(decode.Int32())

	o.Tracef("CALL c'%s' f'%s'", contract.String(), function.String())
	results, err := o.processCallUnlocked(contract, function, params, transfer)
	if err != nil {
		o.Panicf("failed to invoke call: %v", err)
	}
	resultsID := o.GetObjectID(wasmhost.KeyReturn, wasmhost.OBJTYPE_MAP)
	o.host.FindObject(resultsID).(*ScDict).kvStore = results
}

func (o *ScContext) processCallUnlocked(contract, function iscp.Hname, params dict.Dict, transfer colored.Balances) (dict.Dict, error) {
	o.wc.proc.instanceLock.Unlock()
	defer o.wc.proc.instanceLock.Lock()

	if o.wc.ctx != nil {
		return o.wc.ctx.Call(contract, function, params, transfer)
	}
	return o.wc.ctxView.Call(contract, function, params)
}

func (o *ScContext) processDeploy(bytes []byte) {
	decode := wasmlib.NewBytesDecoder(bytes)
	programHash, err := hashing.HashValueFromBytes(decode.Hash().Bytes())
	if err != nil {
		o.Panicf(err.Error())
	}
	name := string(decode.Bytes())
	description := string(decode.Bytes())
	params := o.getParams(decode.Int32())
	o.Tracef("DEPLOY c'%s' f'%s'", name, description)
	err = o.processDeployUnlocked(programHash, name, description, params)
	if err != nil {
		o.Panicf("failed to deploy: %v", err)
	}
}

func (o *ScContext) processDeployUnlocked(programHash hashing.HashValue, name, description string, params dict.Dict) error {
	o.wc.proc.instanceLock.Unlock()
	defer o.wc.proc.instanceLock.Lock()

	return o.wc.ctx.DeployContract(programHash, name, description, params)
}

func (o *ScContext) processPost(bytes []byte) {
	decode := wasmlib.NewBytesDecoder(bytes)
	chainID, err := iscp.ChainIDFromBytes(decode.ChainID().Bytes())
	if err != nil {
		o.Panicf(err.Error())
	}
	contract, err := iscp.HnameFromBytes(decode.Hname().Bytes())
	if err != nil {
		o.Panicf(err.Error())
	}
	function, err := iscp.HnameFromBytes(decode.Hname().Bytes())
	if err != nil {
		o.Panicf(err.Error())
	}
	o.Tracef("POST c'%s' f'%s'", contract.String(), function.String())
	params := o.getParams(decode.Int32())
	transferID := decode.Int32()
	if transferID == 0 {
		o.Panicf("transfer is required for post")
	}
	transfer := o.getTransfer(transferID)
	metadata := &iscp.SendMetadata{
		TargetContract: contract,
		EntryPoint:     function,
		Args:           params,
	}
	delay := decode.Int32()
	if delay == 0 {
		if !o.wc.ctx.Send(chainID.AsAddress(), transfer, metadata) {
			o.Panicf("failed to send to %s", chainID.AsAddress().String())
		}
		return
	}

	if delay < 0 {
		o.Panicf("invalid delay: %d", delay)
	}

	timeLock := time.Unix(0, o.wc.ctx.GetTimestamp())
	timeLock = timeLock.Add(time.Duration(delay) * time.Second)
	options := iscp.SendOptions{
		TimeLock: uint32(timeLock.Unix()),
	}
	if !o.wc.ctx.Send(chainID.AsAddress(), transfer, metadata, options) {
		o.Panicf("failed to send to %s", chainID.AsAddress().String())
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

func (o *ScContext) getTransfer(transferID int32) colored.Balances {
	if transferID == 0 {
		return colored.NewBalances()
	}
	transfer := colored.NewBalances()
	transferDict := o.host.FindObject(transferID).(*ScDict).kvStore
	transferDict.MustIterate("", func(key kv.Key, value []byte) bool {
		col, err := codec.DecodeColor([]byte(key))
		if err != nil {
			o.Panicf(err.Error())
		}
		amount, err := codec.DecodeUint64(value)
		if err != nil {
			o.Panicf(err.Error())
		}
		o.Tracef("  XFER %d '%s'", amount, col.String())
		transfer.Set(col, amount)
		return true
	})
	return transfer
}
