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
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmrequests"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

// NOTE: These functions correspond to the Sandbox fnXxx constants in WasmLib
var sandboxFunctions = []func(*WasmHostSandbox, []byte) []byte{
	nil,
	(*WasmHostSandbox).fnAccountID,
	(*WasmHostSandbox).fnBalance,
	(*WasmHostSandbox).fnBalances,
	(*WasmHostSandbox).fnBlockContext,
	(*WasmHostSandbox).fnCall,
	(*WasmHostSandbox).fnCaller,
	(*WasmHostSandbox).fnChainID,
	(*WasmHostSandbox).fnChainOwnerID,
	(*WasmHostSandbox).fnContract,
	(*WasmHostSandbox).fnContractCreator,
	(*WasmHostSandbox).fnDeployContract,
	(*WasmHostSandbox).fnEntropy,
	(*WasmHostSandbox).fnEvent,
	(*WasmHostSandbox).fnIncomingTransfer,
	(*WasmHostSandbox).fnLog,
	(*WasmHostSandbox).fnMinted,
	(*WasmHostSandbox).fnPanic,
	(*WasmHostSandbox).fnParams,
	(*WasmHostSandbox).fnPost,
	(*WasmHostSandbox).fnRequest,
	(*WasmHostSandbox).fnRequestID,
	(*WasmHostSandbox).fnResults,
	(*WasmHostSandbox).fnSend,
	(*WasmHostSandbox).fnStateAnchor,
	(*WasmHostSandbox).fnTimestamp,
	(*WasmHostSandbox).fnTrace,
	(*WasmHostSandbox).fnUtilsBase58Decode,
	(*WasmHostSandbox).fnUtilsBase58Encode,
	(*WasmHostSandbox).fnUtilsBlsAddress,
	(*WasmHostSandbox).fnUtilsBlsAggregate,
	(*WasmHostSandbox).fnUtilsBlsValid,
	(*WasmHostSandbox).fnUtilsEd25519Address,
	(*WasmHostSandbox).fnUtilsEd25519Valid,
	(*WasmHostSandbox).fnUtilsHashBlake2b,
	(*WasmHostSandbox).fnUtilsHashName,
	(*WasmHostSandbox).fnUtilsHashSha3,
}

// '$' prefix indicates a string param
// '$' prefix indicates a bytes param
// otherwise there is no param
// NOTE: These strings correspond to the Sandbox fnXxx constants in WasmLib
var sandboxFuncNames = []string{
	"nil",
	"FnAccountID",
	"#FnBalance",
	"FnBalances",
	"FnBlockContext",
	"FnCall",
	"FnCaller",
	"FnChainID",
	"FnChainOwnerID",
	"FnContract",
	"FnContractCreator",
	"#FnDeployContract",
	"FnEntropy",
	"$FnEvent",
	"FnIncomingTransfer",
	"$FnLog",
	"FnMinted",
	"$FnPanic",
	"FnParams",
	"#FnPost",
	"FnRequest",
	"FnRequestID",
	"#FnResults",
	"#FnSend",
	"#FnStateAnchor",
	"FnTimestamp",
	"$FnTrace",
	"$FnUtilsBase58Decode",
	"#FnUtilsBase58Encode",
	"#FnUtilsBlsAddress",
	"#FnUtilsBlsAggregate",
	"#FnUtilsBlsValid",
	"#FnUtilsEd25519Address",
	"#FnUtilsEd25519Valid",
	"#FnUtilsHashBlake2b",
	"$FnUtilsHashName",
	"#FnUtilsHashSha3",
}

type WasmHostSandbox struct {
	common  iscp.SandboxBase
	ctx     iscp.Sandbox
	ctxView iscp.SandboxView
	wc      *WasmContext
}

var _ ISandbox = new(WasmHostSandbox)

func NewWasmHostSandbox(wc *WasmContext, ctx interface{}) *WasmHostSandbox {
	s := new(WasmHostSandbox)
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

func (s *WasmHostSandbox) Call(funcNr int32, params []byte) []byte {
	return sandboxFunctions[-funcNr](s, params)
}

func (s *WasmHostSandbox) checkErr(err error) {
	if err != nil {
		s.Panicf(err.Error())
	}
}

func (s *WasmHostSandbox) Panicf(format string, args ...interface{}) {
	s.common.Log().Panicf(format, args...)
}

func (s *WasmHostSandbox) Tracef(format string, args ...interface{}) {
	s.common.Log().Debugf(format, args...)
}

//////////////////// sandbox functions \\\\\\\\\\\\\\\\\\\\

func (s *WasmHostSandbox) fnAccountID(args []byte) []byte {
	return s.common.AccountID().Bytes()
}

func (s *WasmHostSandbox) fnBalance(args []byte) []byte {
	color, err := colored.ColorFromBytes(args)
	s.checkErr(err)
	return codec.EncodeUint64(s.ctx.Balance(color))
}

func (s *WasmHostSandbox) fnBalances(args []byte) []byte {
	return s.common.Balances().Bytes()
}

func (s *WasmHostSandbox) fnBlockContext(args []byte) []byte {
	panic("implement me")
}

func (s *WasmHostSandbox) fnCall(args []byte) []byte {
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

func (s *WasmHostSandbox) callUnlocked(contract, function iscp.Hname, params dict.Dict, transfer colored.Balances) (dict.Dict, error) {
	s.wc.proc.instanceLock.Unlock()
	defer s.wc.proc.instanceLock.Lock()

	if s.ctx != nil {
		return s.ctx.Call(contract, function, params, transfer)
	}
	return s.ctxView.Call(contract, function, params)
}

func (s *WasmHostSandbox) fnCaller(args []byte) []byte {
	return s.ctx.Caller().Bytes()
}

func (s *WasmHostSandbox) fnChainID(args []byte) []byte {
	return s.common.ChainID().Bytes()
}

func (s *WasmHostSandbox) fnChainOwnerID(args []byte) []byte {
	return s.common.ChainOwnerID().Bytes()
}

func (s *WasmHostSandbox) fnContract(args []byte) []byte {
	return s.common.Contract().Bytes()
}

func (s *WasmHostSandbox) fnContractCreator(args []byte) []byte {
	return s.common.ContractCreator().Bytes()
}

func (s *WasmHostSandbox) fnDeployContract(args []byte) []byte {
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

func (s *WasmHostSandbox) deployUnlocked(programHash hashing.HashValue, name, description string, params dict.Dict) error {
	s.wc.proc.instanceLock.Unlock()
	defer s.wc.proc.instanceLock.Lock()

	return s.ctx.DeployContract(programHash, name, description, params)
}

func (s *WasmHostSandbox) fnEntropy(args []byte) []byte {
	return s.ctx.GetEntropy().Bytes()
}

func (s *WasmHostSandbox) fnEvent(args []byte) []byte {
	s.ctx.Event(string(args))
	return nil
}

func (s *WasmHostSandbox) fnIncomingTransfer(args []byte) []byte {
	return s.ctx.IncomingTransfer().Bytes()
}

func (s *WasmHostSandbox) fnLog(args []byte) []byte {
	s.common.Log().Infof(string(args))
	return nil
}

func (s *WasmHostSandbox) fnMinted(args []byte) []byte {
	return s.ctx.Minted().Bytes()
}

func (s *WasmHostSandbox) fnPanic(args []byte) []byte {
	s.common.Log().Panicf("WASM panic: %s", string(args))
	return nil
}

func (s *WasmHostSandbox) fnParams(args []byte) []byte {
	return s.common.Params().Bytes()
}

func (s *WasmHostSandbox) fnPost(args []byte) []byte {
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

func (s *WasmHostSandbox) fnRequest(args []byte) []byte {
	return s.ctx.Request().Bytes()
}

func (s *WasmHostSandbox) fnRequestID(args []byte) []byte {
	return s.ctx.Request().ID().Bytes()
}

func (s *WasmHostSandbox) fnResults(args []byte) []byte {
	results, err := dict.FromBytes(args)
	if err != nil {
		s.Panicf("call results: %s", err.Error())
	}
	s.wc.results = results
	return nil
}

// transfer tokens to address
func (s *WasmHostSandbox) fnSend(args []byte) []byte {
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

func (s *WasmHostSandbox) fnStateAnchor(args []byte) []byte {
	panic("implement me")
}

func (s *WasmHostSandbox) fnTimestamp(args []byte) []byte {
	return codec.EncodeInt64(s.common.GetTimestamp())
}

func (s *WasmHostSandbox) fnTrace(args []byte) []byte {
	s.common.Log().Debugf(string(args))
	return nil
}

func (s WasmHostSandbox) fnUtilsBase58Decode(args []byte) []byte {
	bytes, err := s.common.Utils().Base58().Decode(string(args))
	s.checkErr(err)
	return bytes
}

func (s WasmHostSandbox) fnUtilsBase58Encode(args []byte) []byte {
	return []byte(s.common.Utils().Base58().Encode(args))
}

func (s WasmHostSandbox) fnUtilsBlsAddress(args []byte) []byte {
	address, err := s.common.Utils().BLS().AddressFromPublicKey(args)
	s.checkErr(err)
	return address.Bytes()
}

func (s WasmHostSandbox) fnUtilsBlsAggregate(args []byte) []byte {
	dec := wasmtypes.NewWasmDecoder(args)
	count := wasmtypes.Uint32Decode(dec)
	pubKeysBin := make([][]byte, count)
	for i := uint32(0); i < count; i++ {
		pubKeysBin[i] = dec.Bytes()
	}
	count = wasmtypes.Uint32Decode(dec)
	sigsBin := make([][]byte, count)
	for i := uint32(0); i < count; i++ {
		sigsBin[i] = dec.Bytes()
	}
	pubKeyBin, sigBin, err := s.common.Utils().BLS().AggregateBLSSignatures(pubKeysBin, sigsBin)
	s.checkErr(err)
	return wasmtypes.NewWasmEncoder().Bytes(pubKeyBin).Bytes(sigBin).Buf()
}

func (s WasmHostSandbox) fnUtilsBlsValid(args []byte) []byte {
	dec := wasmtypes.NewWasmDecoder(args)
	data := dec.Bytes()
	pubKey := dec.Bytes()
	signature := dec.Bytes()
	valid := s.common.Utils().BLS().ValidSignature(data, pubKey, signature)
	return codec.EncodeBool(valid)
}

func (s WasmHostSandbox) fnUtilsEd25519Address(args []byte) []byte {
	address, err := s.common.Utils().ED25519().AddressFromPublicKey(args)
	s.checkErr(err)
	return address.Bytes()
}

func (s WasmHostSandbox) fnUtilsEd25519Valid(args []byte) []byte {
	dec := wasmtypes.NewWasmDecoder(args)
	data := dec.Bytes()
	pubKey := dec.Bytes()
	signature := dec.Bytes()
	valid := s.common.Utils().ED25519().ValidSignature(data, pubKey, signature)
	return codec.EncodeBool(valid)
}

func (s WasmHostSandbox) fnUtilsHashBlake2b(args []byte) []byte {
	return s.common.Utils().Hashing().Blake2b(args).Bytes()
}

func (s WasmHostSandbox) fnUtilsHashName(args []byte) []byte {
	return codec.EncodeHname(s.common.Utils().Hashing().Hname(string(args)))
}

func (s WasmHostSandbox) fnUtilsHashSha3(args []byte) []byte {
	return s.common.Utils().Hashing().Sha3(args).Bytes()
}
