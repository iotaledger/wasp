// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"time"

	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmrequests"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

// NOTE: These functions correspond to the Sandbox fnXxx constants in WasmLib
var sandboxFunctions = []func(*WasmContextSandbox, []byte) []byte{
	nil,
	(*WasmContextSandbox).fnAccountID,
	(*WasmContextSandbox).fnAllowance,
	(*WasmContextSandbox).fnBalance,
	(*WasmContextSandbox).fnBalances,
	(*WasmContextSandbox).fnBlockContext,
	(*WasmContextSandbox).fnCall,
	(*WasmContextSandbox).fnCaller,
	(*WasmContextSandbox).fnChainID,
	(*WasmContextSandbox).fnChainOwnerID,
	(*WasmContextSandbox).fnContract,
	(*WasmContextSandbox).fnContractCreator,
	(*WasmContextSandbox).fnDeployContract,
	(*WasmContextSandbox).fnEntropy,
	(*WasmContextSandbox).fnEvent,
	(*WasmContextSandbox).fnLog,
	(*WasmContextSandbox).fnMinted,
	(*WasmContextSandbox).fnPanic,
	(*WasmContextSandbox).fnParams,
	(*WasmContextSandbox).fnPost,
	(*WasmContextSandbox).fnRequest,
	(*WasmContextSandbox).fnRequestID,
	(*WasmContextSandbox).fnResults,
	(*WasmContextSandbox).fnSend,
	(*WasmContextSandbox).fnStateAnchor,
	(*WasmContextSandbox).fnTimestamp,
	(*WasmContextSandbox).fnTrace,
	(*WasmContextSandbox).fnUtilsBase58Decode,
	(*WasmContextSandbox).fnUtilsBase58Encode,
	(*WasmContextSandbox).fnUtilsBlsAddress,
	(*WasmContextSandbox).fnUtilsBlsAggregate,
	(*WasmContextSandbox).fnUtilsBlsValid,
	(*WasmContextSandbox).fnUtilsEd25519Address,
	(*WasmContextSandbox).fnUtilsEd25519Valid,
	(*WasmContextSandbox).fnUtilsHashBlake2b,
	(*WasmContextSandbox).fnUtilsHashName,
	(*WasmContextSandbox).fnUtilsHashSha3,
	(*WasmContextSandbox).fnTransferAllowed,
	(*WasmContextSandbox).fnEstimateDust,
}

// '$' prefix indicates a string param
// '#' prefix indicates a bytes param
// otherwise there is no param
// NOTE: These strings correspond to the Sandbox fnXxx constants in WasmLib
var sandboxFuncNames = []string{
	"nil",
	"FnAccountID",
	"FnAllowance",
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
	"#FnTransferAllowed",
	"#FnEstimateDust",
}

// WasmContextSandbox is the host side of the WasmLib Sandbox interface
// It acts as a change-resistant layer to wrap changes to the ISCP sandbox,
// to limit bothering users of WasmLib as little as possible with those changes.
type WasmContextSandbox struct {
	common  iscp.SandboxBase
	ctx     iscp.Sandbox
	ctxView iscp.SandboxView
	cvt     WasmConvertor
	wc      *WasmContext
}

var _ ISandbox = new(WasmContextSandbox)

func NewWasmContextSandbox(wc *WasmContext, ctx interface{}) *WasmContextSandbox {
	s := &WasmContextSandbox{wc: wc}
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

func (s *WasmContextSandbox) Call(funcNr int32, params []byte) []byte {
	return sandboxFunctions[-funcNr](s, params)
}

func (s *WasmContextSandbox) checkErr(err error) {
	if err != nil {
		s.Panicf(err.Error())
	}
}

func (s *WasmContextSandbox) makeRequest(args []byte) iscp.RequestParameters {
	req := wasmrequests.NewPostRequestFromBytes(args)
	chainID := s.cvt.IscpChainID(&req.ChainID)
	contract := s.cvt.IscpHname(req.Contract)
	function := s.cvt.IscpHname(req.Function)
	params, err := dict.FromBytes(req.Params)
	s.checkErr(err)

	scAssets := wasmlib.NewScAssets(req.Transfer)
	allowance := s.cvt.IscpAllowance(scAssets)
	assets := allowance
	// Force a minimum transfer of 1000 iotas for dust and some gas
	// excess can always be reclaimed from the chain account by the user
	// This also removes the silly requirement to transfer 1 iota
	if assets.Assets.Iotas < 1000 {
		// assets are different from allowance, so clone allowance before modifying
		assets = allowance.Clone()
		assets.Assets.Iotas = 1000
	}

	s.Tracef("POST %s.%s, chain %s", contract.String(), function.String(), chainID.String())
	sendReq := iscp.RequestParameters{
		AdjustToMinimumDustDeposit: true,
		TargetAddress:              chainID.AsAddress(),
		FungibleTokens:             assets.Assets,
		Metadata: &iscp.SendMetadata{
			TargetContract: contract,
			EntryPoint:     function,
			Params:         params,
			// TODO check, probably not correct
			Allowance: allowance,
			GasBudget: 1_000_000,
		},
	}
	if req.Delay != 0 {
		timeLock := time.Unix(0, s.ctx.Timestamp())
		timeLock = timeLock.Add(time.Duration(req.Delay) * time.Second)
		sendReq.Options.Timelock = &iscp.TimeData{Time: timeLock}
	}
	return sendReq
}

func (s *WasmContextSandbox) Panicf(format string, args ...interface{}) {
	s.common.Log().Panicf(format, args...)
}

func (s *WasmContextSandbox) Tracef(format string, args ...interface{}) {
	s.common.Log().Debugf(format, args...)
}

//////////////////// sandbox functions \\\\\\\\\\\\\\\\\\\\

func (s *WasmContextSandbox) fnAccountID(args []byte) []byte {
	return s.cvt.ScAgentID(s.common.AccountID()).Bytes()
}

func (s *WasmContextSandbox) fnAllowance(args []byte) []byte {
	allowance := s.ctx.AllowanceAvailable()
	return s.cvt.ScBalances(allowance).Bytes()
}

func (s *WasmContextSandbox) fnBalance(args []byte) []byte {
	if len(args) == 0 {
		return codec.EncodeUint64(s.common.BalanceIotas())
	}
	tokenID := wasmtypes.TokenIDFromBytes(args)
	token := s.cvt.IscpTokenID(&tokenID)
	return codec.EncodeUint64(s.common.BalanceNativeToken(token).Uint64())
}

func (s *WasmContextSandbox) fnBalances(args []byte) []byte {
	allowance := &iscp.Allowance{}
	allowance.Assets = s.common.BalanceFungibleTokens()
	allowance.NFTs = s.common.OwnedNFTs()
	return s.cvt.ScBalances(allowance).Bytes()
}

func (s *WasmContextSandbox) fnBlockContext(args []byte) []byte {
	panic("implement me")
}

func (s *WasmContextSandbox) fnCall(args []byte) []byte {
	req := wasmrequests.NewCallRequestFromBytes(args)
	contract := s.cvt.IscpHname(req.Contract)
	function := s.cvt.IscpHname(req.Function)
	params, err := dict.FromBytes(req.Params)
	s.checkErr(err)
	scAssets := wasmlib.NewScAssets(req.Transfer)
	allowance := s.cvt.IscpAllowance(scAssets)
	// TODO check, probably not right
	transfer := iscp.NewAllowanceFungibleTokens(allowance.Assets)
	s.Tracef("CALL %s.%s", contract.String(), function.String())
	results := s.callUnlocked(contract, function, params, transfer)
	return results.Bytes()
}

func (s *WasmContextSandbox) callUnlocked(contract, function iscp.Hname, params dict.Dict, transfer *iscp.Allowance) dict.Dict {
	s.wc.proc.instanceLock.Unlock()
	defer s.wc.proc.instanceLock.Lock()

	if s.ctx != nil {
		return s.ctx.Call(contract, function, params, transfer)
	}
	return s.ctxView.Call(contract, function, params)
}

func (s *WasmContextSandbox) fnCaller(args []byte) []byte {
	return s.cvt.ScAgentID(s.ctx.Caller()).Bytes()
}

func (s *WasmContextSandbox) fnChainID(args []byte) []byte {
	return s.cvt.ScChainID(s.common.ChainID()).Bytes()
}

func (s *WasmContextSandbox) fnChainOwnerID(args []byte) []byte {
	return s.cvt.ScAgentID(s.common.ChainOwnerID()).Bytes()
}

func (s *WasmContextSandbox) fnContract(args []byte) []byte {
	return s.cvt.ScHname(s.common.Contract()).Bytes()
}

func (s *WasmContextSandbox) fnContractCreator(args []byte) []byte {
	return s.cvt.ScAgentID(s.common.ContractCreator()).Bytes()
}

func (s *WasmContextSandbox) fnDeployContract(args []byte) []byte {
	req := wasmrequests.NewDeployRequestFromBytes(args)
	programHash, err := hashing.HashValueFromBytes(req.ProgHash.Bytes())
	s.checkErr(err)
	initParams, err := dict.FromBytes(req.Params)
	s.checkErr(err)
	s.Tracef("DEPLOY %s: %s", req.Name, req.Description)
	s.deployUnlocked(programHash, req.Name, req.Description, initParams)
	return nil
}

func (s *WasmContextSandbox) deployUnlocked(programHash hashing.HashValue, name, description string, params dict.Dict) {
	s.wc.proc.instanceLock.Unlock()
	defer s.wc.proc.instanceLock.Lock()

	s.ctx.DeployContract(programHash, name, description, params)
}

func (s *WasmContextSandbox) fnEntropy(args []byte) []byte {
	return s.cvt.ScHash(s.ctx.GetEntropy()).Bytes()
}

func (s *WasmContextSandbox) fnEstimateDust(args []byte) []byte {
	dust := s.ctx.EstimateRequiredDustDeposit(s.makeRequest(args))
	return codec.EncodeUint64(dust)
}

func (s *WasmContextSandbox) fnEvent(args []byte) []byte {
	s.ctx.Event(string(args))
	return nil
}

func (s *WasmContextSandbox) fnLog(args []byte) []byte {
	s.common.Log().Infof(string(args))
	return nil
}

func (s *WasmContextSandbox) fnMinted(args []byte) []byte {
	panic("fixme: wc.fnMinted")
	// return s.ctx.Minted().Bytes()
}

func (s *WasmContextSandbox) fnPanic(args []byte) []byte {
	s.common.Log().Panicf("WASM: panic in VM: %s", string(args))
	return nil
}

func (s *WasmContextSandbox) fnParams(args []byte) []byte {
	return s.common.Params().Dict.Bytes()
}

func (s *WasmContextSandbox) fnPost(args []byte) []byte {
	s.ctx.Send(s.makeRequest(args))
	return nil
}

func (s *WasmContextSandbox) fnRequest(args []byte) []byte {
	panic("fixme: wc.fnRequest")
	// return s.ctx.Request().Bytes()
}

func (s *WasmContextSandbox) fnRequestID(args []byte) []byte {
	return s.cvt.ScRequestID(s.ctx.Request().ID()).Bytes()
}

func (s *WasmContextSandbox) fnResults(args []byte) []byte {
	results, err := dict.FromBytes(args)
	if err != nil {
		s.Panicf("call results: %s", err.Error())
	}
	s.wc.results = results
	return nil
}

// transfer tokens to address
func (s *WasmContextSandbox) fnSend(args []byte) []byte {
	req := wasmrequests.NewSendRequestFromBytes(args)
	address := s.cvt.IscpAddress(&req.Address)
	scAssets := wasmlib.NewScAssets(req.Transfer)
	if !scAssets.IsEmpty() {
		allowance := s.cvt.IscpAllowance(scAssets)
		s.ctx.Send(iscp.RequestParameters{
			AdjustToMinimumDustDeposit: true,
			TargetAddress:              address,
			FungibleTokens:             allowance.Assets,
		})
	}
	return nil
}

func (s *WasmContextSandbox) fnStateAnchor(args []byte) []byte {
	panic("implement me")
}

func (s *WasmContextSandbox) fnTimestamp(args []byte) []byte {
	return codec.EncodeInt64(s.common.Timestamp())
}

func (s *WasmContextSandbox) fnTrace(args []byte) []byte {
	s.common.Log().Debugf(string(args))
	return nil
}

// transfer tokens to address
func (s *WasmContextSandbox) fnTransferAllowed(args []byte) []byte {
	req := wasmrequests.NewTransferRequestFromBytes(args)
	agentID := s.cvt.IscpAgentID(&req.AgentID)
	scAssets := wasmlib.NewScAssets(req.Transfer)
	if !scAssets.IsEmpty() {
		allowance := s.cvt.IscpAllowance(scAssets)
		if req.Create {
			s.ctx.TransferAllowedFundsForceCreateTarget(agentID, allowance)
		} else {
			s.ctx.TransferAllowedFunds(agentID, allowance)
		}
	}
	return nil
}

func (s WasmContextSandbox) fnUtilsBase58Decode(args []byte) []byte {
	bytes, err := s.common.Utils().Base58().Decode(string(args))
	s.checkErr(err)
	return bytes
}

func (s WasmContextSandbox) fnUtilsBase58Encode(args []byte) []byte {
	return []byte(s.common.Utils().Base58().Encode(args))
}

func (s WasmContextSandbox) fnUtilsBlsAddress(args []byte) []byte {
	address, err := s.common.Utils().BLS().AddressFromPublicKey(args)
	s.checkErr(err)
	return s.cvt.ScAddress(address).Bytes()
}

func (s WasmContextSandbox) fnUtilsBlsAggregate(args []byte) []byte {
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

func (s WasmContextSandbox) fnUtilsBlsValid(args []byte) []byte {
	dec := wasmtypes.NewWasmDecoder(args)
	data := dec.Bytes()
	pubKey := dec.Bytes()
	signature := dec.Bytes()
	valid := s.common.Utils().BLS().ValidSignature(data, pubKey, signature)
	return codec.EncodeBool(valid)
}

func (s WasmContextSandbox) fnUtilsEd25519Address(args []byte) []byte {
	address, err := s.common.Utils().ED25519().AddressFromPublicKey(args)
	s.checkErr(err)
	return s.cvt.ScAddress(address).Bytes()
}

func (s WasmContextSandbox) fnUtilsEd25519Valid(args []byte) []byte {
	dec := wasmtypes.NewWasmDecoder(args)
	data := dec.Bytes()
	pubKey := dec.Bytes()
	signature := dec.Bytes()
	valid := s.common.Utils().ED25519().ValidSignature(data, pubKey, signature)
	return codec.EncodeBool(valid)
}

func (s WasmContextSandbox) fnUtilsHashBlake2b(args []byte) []byte {
	return s.cvt.ScHash(s.common.Utils().Hashing().Blake2b(args)).Bytes()
}

func (s WasmContextSandbox) fnUtilsHashName(args []byte) []byte {
	return s.cvt.ScHname(s.common.Utils().Hashing().Hname(string(args))).Bytes()
}

func (s WasmContextSandbox) fnUtilsHashSha3(args []byte) []byte {
	return s.cvt.ScHash(s.common.Utils().Hashing().Sha3(args)).Bytes()
}
