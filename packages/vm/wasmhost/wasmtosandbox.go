// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/wasmlib/go/wasmlib/wasmrequests"
)

// NOTE: These functions correspond to the Sandbox fnXxx constants in WasmLib
var sandboxFunctions = []func(*WasmToSandbox, []byte) []byte{
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
	wc      *WasmContext
}

var _ ISandbox = new(WasmToSandbox)

func NewWasmToSandbox(wc *WasmContext, ctx interface{}) *WasmToSandbox {
	s := new(WasmToSandbox)
	s.wc = wc
	switch tctx := ctx.(type) {
	case iscp.Sandbox:
		s.common = tctx
		s.ctx = tctx
	case iscp.SandboxView:
		s.common = tctx
		s.ctxView = tctx
	default:
		panic(iscp.ErrWrongTypeEntryPoint)
	}
	return s
}

func (s *WasmToSandbox) Call(funcNr int32, params []byte) []byte {
	return sandboxFunctions[-funcNr](s, params)
}

func (s *WasmToSandbox) checkErr(err error) {
	if err != nil {
		s.Panicf(err.Error())
	}
}

func (s *WasmToSandbox) Panicf(format string, args ...interface{}) {
	s.common.Log().Panicf(format, args...)
}

func (s *WasmToSandbox) Tracef(format string, args ...interface{}) {
	s.common.Log().Debugf(format, args...)
}

//////////////////// sandbox functions \\\\\\\\\\\\\\\\\\\\

func (s *WasmToSandbox) fnAccountID(args []byte) []byte {
	return s.common.AccountID().Bytes()
}

func (s *WasmToSandbox) fnBalance(args []byte) []byte {
	color, err := colored.ColorFromBytes(args)
	s.checkErr(err)
	return codec.EncodeUint64(s.ctx.Balance(color))
}

func (s *WasmToSandbox) fnBalances(args []byte) []byte {
	return s.common.Balances().Bytes()
}

func (s *WasmToSandbox) fnBlockContext(args []byte) []byte {
	panic("implement me")
}

func (s *WasmToSandbox) fnCall(args []byte) []byte {
	req := wasmrequests.NewCallRequestFromBytes(args)
	contract, err := iscp.HnameFromBytes(req.Contract.Bytes())
	s.checkErr(err)
	function, err := iscp.HnameFromBytes(req.Function.Bytes())
	s.checkErr(err)
	params, err := dict.FromBytes(req.Params)
	s.checkErr(err)
	transfer, err := colored.BalancesFromBytes(req.Transfer)
	s.checkErr(err)
	s.Tracef("CALL hContract '%s, hFunction %s", contract.String(), function.String())
	results, err := s.callUnlocked(contract, function, params, transfer)
	s.checkErr(err)
	return results.Bytes()
}

func (s *WasmToSandbox) callUnlocked(contract, function iscp.Hname, params dict.Dict, transfer colored.Balances) (dict.Dict, error) {
	s.wc.proc.instanceLock.Unlock()
	defer s.wc.proc.instanceLock.Lock()

	if s.ctx != nil {
		return s.ctx.Call(contract, function, params, transfer)
	}
	return s.ctxView.Call(contract, function, params)
}

func (s *WasmToSandbox) fnCaller(args []byte) []byte {
	return s.ctx.Caller().Bytes()
}

func (s *WasmToSandbox) fnChainID(args []byte) []byte {
	return s.common.ChainID().Bytes()
}

func (s *WasmToSandbox) fnChainOwnerID(args []byte) []byte {
	return s.common.ChainOwnerID().Bytes()
}

func (s *WasmToSandbox) fnContract(args []byte) []byte {
	return s.common.Contract().Bytes()
}

func (s *WasmToSandbox) fnContractCreator(args []byte) []byte {
	return s.common.ContractCreator().Bytes()
}

func (s *WasmToSandbox) fnDeployContract(args []byte) []byte {
	req := wasmrequests.NewDeployRequestFromBytes(args)
	programHash, err := hashing.HashValueFromBytes(req.ProgHash.Bytes())
	s.checkErr(err)
	initParams, err := dict.FromBytes(req.Params)
	s.checkErr(err)
	s.Tracef("DEPLOY %s: %s", req.Name, req.Description)
	err = s.deployUnlocked(programHash, req.Name, req.Description, initParams)
	s.checkErr(err)
	return nil
}

func (s *WasmToSandbox) deployUnlocked(programHash hashing.HashValue, name, description string, params dict.Dict) error {
	s.wc.proc.instanceLock.Unlock()
	defer s.wc.proc.instanceLock.Lock()

	return s.ctx.DeployContract(programHash, name, description, params)
}

func (s *WasmToSandbox) fnEntropy(args []byte) []byte {
	return s.ctx.GetEntropy().Bytes()
}

func (s *WasmToSandbox) fnEvent(args []byte) []byte {
	s.ctx.Event(string(args))
	return nil
}

func (s *WasmToSandbox) fnIncomingTransfer(args []byte) []byte {
	return s.ctx.IncomingTransfer().Bytes()
}

func (s *WasmToSandbox) fnLog(args []byte) []byte {
	s.common.Log().Infof(string(args))
	return nil
}

func (s *WasmToSandbox) fnMinted(args []byte) []byte {
	return s.ctx.Minted().Bytes()
}

func (s *WasmToSandbox) fnPanic(args []byte) []byte {
	s.common.Log().Panicf(string(args))
	return nil
}

func (s *WasmToSandbox) fnParams(args []byte) []byte {
	return s.common.Params().Bytes()
}

func (s *WasmToSandbox) fnPost(args []byte) []byte {
	req := wasmrequests.NewPostRequestFromBytes(args)
	chainID, err := iscp.ChainIDFromBytes(req.ChainID.Bytes())
	s.checkErr(err)
	contract, err := iscp.HnameFromBytes(req.Contract.Bytes())
	s.checkErr(err)
	function, err := iscp.HnameFromBytes(req.Function.Bytes())
	s.checkErr(err)
	params, err := dict.FromBytes(req.Params)
	s.checkErr(err)
	transfer, err := colored.BalancesFromBytes(req.Transfer)
	s.checkErr(err)
	if len(transfer) == 0 {
		s.Panicf("transfer is required for post")
	}

	s.Tracef("POST hContract '%s, hFunction %s, chain ", contract.String(), function.String(), chainID.String())
	metadata := &iscp.SendMetadata{
		TargetContract: contract,
		EntryPoint:     function,
		Args:           params,
	}
	if req.Delay == 0 {
		if !s.ctx.Send(chainID.AsAddress(), transfer, metadata) {
			s.Panicf("failed to send to %s", chainID.AsAddress().String())
		}
		return nil
	}

	timeLock := time.Unix(0, s.ctx.GetTimestamp())
	timeLock = timeLock.Add(time.Duration(req.Delay) * time.Second)
	options := iscp.SendOptions{
		TimeLock: uint32(timeLock.Unix()),
	}
	if !s.ctx.Send(chainID.AsAddress(), transfer, metadata, options) {
		s.Panicf("failed to send to %s", chainID.AsAddress().String())
	}
	return nil
}

func (s *WasmToSandbox) fnRequest(args []byte) []byte {
	return s.ctx.Request().Bytes()
}

func (s *WasmToSandbox) fnRequestID(args []byte) []byte {
	return s.ctx.Request().ID().Bytes()
}

// transfer tokens to address
func (s *WasmToSandbox) fnSend(args []byte) []byte {
	req := wasmrequests.NewSendRequestFromBytes(args)
	address, _, err := ledgerstate.AddressFromBytes(req.Address.Bytes())
	s.checkErr(err)
	transfer, err := colored.BalancesFromBytes(req.Transfer)
	s.checkErr(err)
	if len(transfer) != 0 {
		if !s.ctx.Send(address, transfer, nil) {
			s.Panicf("failed to send to %s", address.String())
		}
	}
	return nil
}

func (s *WasmToSandbox) fnStateAnchor(args []byte) []byte {
	panic("implement me")
}

func (s *WasmToSandbox) fnTimestamp(args []byte) []byte {
	return codec.EncodeInt64(s.common.GetTimestamp())
}

func (s *WasmToSandbox) fnTrace(args []byte) []byte {
	s.common.Log().Debugf(string(args))
	return nil
}
