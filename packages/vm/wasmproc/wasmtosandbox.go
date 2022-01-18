// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmproc

import (
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib"
)

const (
	FnAccountID           = int32(-1)
	FnBalance             = int32(-2)
	FnBalances            = int32(-3)
	FnBlockContext        = int32(-4)
	FnCall                = int32(-5)
	FnCaller              = int32(-6)
	FnChainID             = int32(-7)
	FnChainOwnerID        = int32(-8)
	FnContract            = int32(-9)
	FnContractCreator     = int32(-10)
	FnDeployContract      = int32(-11)
	FnEntropy             = int32(-12)
	FnEvent               = int32(-13)
	FnIncomingTransfer    = int32(-14)
	FnLog                 = int32(-15)
	FnMinted              = int32(-16)
	FnPanic               = int32(-17)
	FnParams              = int32(-18)
	FnPost                = int32(-19)
	FnRequest             = int32(-20)
	FnRequestID           = int32(-21)
	FnSend                = int32(-22)
	FnStateAnchor         = int32(-23)
	FnTimestamp           = int32(-24)
	FnTrace               = int32(-25)
	FnUtilsBase58Decode   = int32(-26)
	FnUtilsBase58Encode   = int32(-27)
	FnUtilsBlsAddress     = int32(-28)
	FnUtilsBlsAggregate   = int32(-29)
	FnUtilsBlsValid       = int32(-30)
	FnUtilsEd25519Address = int32(-31)
	FnUtilsEd25519Valid   = int32(-32)
	FnUtilsHashBlake2b    = int32(-33)
	FnUtilsHashName       = int32(-34)
	FnUtilsHashSha3       = int32(-35)
	FnZzzLastItem         = int32(-36)
)

var sandboxFunctions = [-FnZzzLastItem]func(*WasmToSandbox, []byte) []byte{
	nil,
	(*WasmToSandbox).fnAccountID,
	(*WasmToSandbox).fnBalance,
	(*WasmToSandbox).fnBalances,
	(*WasmToSandbox).fnBlockContext,
	(*WasmToSandbox).fnCall,
	(*WasmToSandbox).fnCaller,
	(*WasmToSandbox).fnChainID,
	(*WasmToSandbox).fnChainOwnerID,
	(*WasmToSandbox).fnContract,
	(*WasmToSandbox).fnContractCreator,
	(*WasmToSandbox).fnDeployContract,
	(*WasmToSandbox).fnEntropy,
	(*WasmToSandbox).fnEvent,
	(*WasmToSandbox).fnIncomingTransfer,
	(*WasmToSandbox).fnLog,
	(*WasmToSandbox).fnMinted,
	(*WasmToSandbox).fnPanic,
	(*WasmToSandbox).fnParams,
	(*WasmToSandbox).fnPost,
	(*WasmToSandbox).fnRequest,
	(*WasmToSandbox).fnRequestID,
	(*WasmToSandbox).fnSend,
	(*WasmToSandbox).fnStateAnchor,
	(*WasmToSandbox).fnTimestamp,
	(*WasmToSandbox).fnTrace,
	(*WasmToSandbox).fnUtilsBase58Decode,
	(*WasmToSandbox).fnUtilsBase58Encode,
	(*WasmToSandbox).fnUtilsBlsAddress,
	(*WasmToSandbox).fnUtilsBlsAggregate,
	(*WasmToSandbox).fnUtilsBlsValid,
	(*WasmToSandbox).fnUtilsEd25519Address,
	(*WasmToSandbox).fnUtilsEd25519Valid,
	(*WasmToSandbox).fnUtilsHashBlake2b,
	(*WasmToSandbox).fnUtilsHashName,
	(*WasmToSandbox).fnUtilsHashSha3,
}

type WasmToSandbox struct {
	common  iscp.SandboxBase
	ctx     iscp.Sandbox
	ctxView iscp.SandboxView
	wc      WasmContext
}

func (f *WasmToSandbox) Call(funcNr int32, args []byte) []byte {
	return sandboxFunctions[-funcNr](f, args)
}

func (f *WasmToSandbox) checkErr(err error) {
	if err != nil {
		f.Panicf(err.Error())
	}
}

func (f *WasmToSandbox) Panicf(format string, args ...interface{}) {
	f.common.Log().Panicf(format, args...)
}

func (f *WasmToSandbox) Tracef(format string, args ...interface{}) {
	f.common.Log().Debugf(format, args...)
}

//////////////////// sandbox functions \\\\\\\\\\\\\\\\\\\\

func (f *WasmToSandbox) fnAccountID(args []byte) []byte {
	return f.common.AccountID().Bytes()
}

func (f *WasmToSandbox) fnBalance(args []byte) []byte {
	color, err := colored.ColorFromBytes(args)
	f.checkErr(err)
	return codec.EncodeUint64(f.ctx.Balance(color))
}

func (f *WasmToSandbox) fnBalances(args []byte) []byte {
	return f.common.Balances().Bytes()
}

func (f *WasmToSandbox) fnBlockContext(args []byte) []byte {
	panic("implement me")
}

func (f *WasmToSandbox) fnCall(args []byte) []byte {
	decode := wasmlib.NewBytesDecoder(args)
	contract, err := iscp.HnameFromBytes(decode.Hname().Bytes())
	f.checkErr(err)
	function, err := iscp.HnameFromBytes(decode.Hname().Bytes())
	f.checkErr(err)
	params, err := dict.FromBytes(decode.Bytes())
	f.checkErr(err)
	transfer, err := colored.BalancesFromBytes(decode.Bytes())
	f.checkErr(err)
	f.Tracef("CALL c'%s' f'%s'", contract.String(), function.String())
	results, err := f.callUnlocked(contract, function, params, transfer)
	f.checkErr(err)
	return results.Bytes()
}

func (f *WasmToSandbox) callUnlocked(contract, function iscp.Hname, params dict.Dict, transfer colored.Balances) (dict.Dict, error) {
	f.wc.proc.instanceLock.Unlock()
	defer f.wc.proc.instanceLock.Lock()

	if f.ctx != nil {
		return f.ctx.Call(contract, function, params, transfer)
	}
	return f.ctxView.Call(contract, function, params)
}

func (f *WasmToSandbox) fnCaller(args []byte) []byte {
	return f.ctx.Caller().Bytes()
}

func (f *WasmToSandbox) fnChainID(args []byte) []byte {
	return f.common.ChainID().Bytes()
}

func (f *WasmToSandbox) fnChainOwnerID(args []byte) []byte {
	return f.common.ChainOwnerID().Bytes()
}

func (f *WasmToSandbox) fnContract(args []byte) []byte {
	return f.common.Contract().Bytes()
}

func (f *WasmToSandbox) fnContractCreator(args []byte) []byte {
	return f.common.ContractCreator().Bytes()
}

func (f *WasmToSandbox) fnDeployContract(args []byte) []byte {
	decode := wasmlib.NewBytesDecoder(args)
	programHash, err := hashing.HashValueFromBytes(decode.Hash().Bytes())
	f.checkErr(err)
	name := string(decode.Bytes())
	description := string(decode.Bytes())
	initParams, err := dict.FromBytes(decode.Bytes())
	f.checkErr(err)
	f.Tracef("DEPLOY c'%s' f'%s'", name, description)
	err = f.deployUnlocked(programHash, name, description, initParams)
	f.checkErr(err)
	return nil
}

func (f *WasmToSandbox) deployUnlocked(programHash hashing.HashValue, name, description string, params dict.Dict) error {
	f.wc.proc.instanceLock.Unlock()
	defer f.wc.proc.instanceLock.Lock()

	return f.ctx.DeployContract(programHash, name, description, params)
}

func (f *WasmToSandbox) fnEntropy(args []byte) []byte {
	return f.ctx.GetEntropy().Bytes()
}

func (f *WasmToSandbox) fnEvent(args []byte) []byte {
	f.ctx.Event(string(args))
	return nil
}

func (f *WasmToSandbox) fnIncomingTransfer(args []byte) []byte {
	return f.ctx.IncomingTransfer().Bytes()
}

func (f *WasmToSandbox) fnLog(args []byte) []byte {
	f.common.Log().Infof(string(args))
	return nil
}

func (f *WasmToSandbox) fnMinted(args []byte) []byte {
	return f.ctx.Minted().Bytes()
}

func (f *WasmToSandbox) fnPanic(args []byte) []byte {
	f.common.Log().Panicf(string(args))
	return nil
}

func (f *WasmToSandbox) fnParams(args []byte) []byte {
	return f.common.Params().Bytes()
}

func (f *WasmToSandbox) fnPost(args []byte) []byte {
	decode := wasmlib.NewBytesDecoder(args)
	chainID, err := iscp.ChainIDFromBytes(decode.ChainID().Bytes())
	f.checkErr(err)
	contract, err := iscp.HnameFromBytes(decode.Hname().Bytes())
	f.checkErr(err)
	function, err := iscp.HnameFromBytes(decode.Hname().Bytes())
	f.checkErr(err)
	f.Tracef("POST c'%s' f'%s'", contract.String(), function.String())
	params, err := dict.FromBytes(decode.Bytes())
	f.checkErr(err)
	transfer, err := colored.BalancesFromBytes(decode.Bytes())
	f.checkErr(err)
	if len(transfer) == 0 {
		f.Panicf("transfer is required for post")
	}
	metadata := &iscp.SendMetadata{
		TargetContract: contract,
		EntryPoint:     function,
		Args:           params,
	}
	delay := decode.Uint32()
	if delay == 0 {
		if !f.ctx.Send(chainID.AsAddress(), transfer, metadata) {
			f.Panicf("failed to send to %s", chainID.AsAddress().String())
		}
		return nil
	}

	timeLock := time.Unix(0, f.ctx.GetTimestamp())
	timeLock = timeLock.Add(time.Duration(delay) * time.Second)
	options := iscp.SendOptions{
		TimeLock: uint32(timeLock.Unix()),
	}
	if !f.ctx.Send(chainID.AsAddress(), transfer, metadata, options) {
		f.Panicf("failed to send to %s", chainID.AsAddress().String())
	}
	return nil
}

func (f *WasmToSandbox) fnRequest(args []byte) []byte {
	return f.ctx.Request().Bytes()
}

func (f *WasmToSandbox) fnRequestID(args []byte) []byte {
	return f.ctx.Request().ID().Bytes()
}

// transfer tokens to address
func (f *WasmToSandbox) fnSend(args []byte) []byte {
	decode := wasmlib.NewBytesDecoder(args)
	address, _, err := ledgerstate.AddressFromBytes(decode.Address().Bytes())
	f.checkErr(err)
	transfer, err := colored.BalancesFromBytes(decode.Bytes())
	f.checkErr(err)
	if len(transfer) != 0 {
		if !f.ctx.Send(address, transfer, nil) {
			f.Panicf("failed to send to %s", address.String())
		}
	}
	return nil
}

func (f *WasmToSandbox) fnStateAnchor(args []byte) []byte {
	return nil
}

func (f *WasmToSandbox) fnTimestamp(args []byte) []byte {
	return codec.EncodeInt64(f.common.GetTimestamp())
}

func (f *WasmToSandbox) fnTrace(args []byte) []byte {
	f.common.Log().Debugf(string(args))
	return nil
}
