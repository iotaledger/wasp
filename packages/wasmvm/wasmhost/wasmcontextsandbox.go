// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmhost

import (
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/vm/gas"
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
	(*WasmContextSandbox).fnDeployContract,
	(*WasmContextSandbox).fnEntropy,
	(*WasmContextSandbox).fnEstimateDust,
	(*WasmContextSandbox).fnEvent,
	(*WasmContextSandbox).fnLog,
	(*WasmContextSandbox).fnMinted,
	(*WasmContextSandbox).fnPanic,
	(*WasmContextSandbox).fnParams,
	(*WasmContextSandbox).fnPost,
	(*WasmContextSandbox).fnRequest,
	(*WasmContextSandbox).fnRequestID,
	(*WasmContextSandbox).fnRequestSender,
	(*WasmContextSandbox).fnResults,
	(*WasmContextSandbox).fnSend,
	(*WasmContextSandbox).fnStateAnchor,
	(*WasmContextSandbox).fnTimestamp,
	(*WasmContextSandbox).fnTrace,
	(*WasmContextSandbox).fnTransferAllowed,
	(*WasmContextSandbox).fnUtilsBech32Decode,
	(*WasmContextSandbox).fnUtilsBech32Encode,
	(*WasmContextSandbox).fnUtilsBlsAddress,
	(*WasmContextSandbox).fnUtilsBlsAggregate,
	(*WasmContextSandbox).fnUtilsBlsValid,
	(*WasmContextSandbox).fnUtilsEd25519Address,
	(*WasmContextSandbox).fnUtilsEd25519Valid,
	(*WasmContextSandbox).fnUtilsHashBlake2b,
	(*WasmContextSandbox).fnUtilsHashName,
	(*WasmContextSandbox).fnUtilsHashSha3,
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
	"#FnDeployContract",
	"FnEntropy",
	"#FnEstimateDust",
	"$FnEvent",
	"$FnLog",
	"FnMinted",
	"$FnPanic",
	"FnParams",
	"#FnPost",
	"FnRequest",
	"FnRequestID",
	"FnRequestSender",
	"#FnResults",
	"#FnSend",
	"#FnStateAnchor",
	"FnTimestamp",
	"$FnTrace",
	"#FnTransferAllowed",
	"$FnUtilsBech32Decode",
	"#FnUtilsBech32Encode",
	"#FnUtilsBlsAddress",
	"#FnUtilsBlsAggregate",
	"#FnUtilsBlsValid",
	"#FnUtilsEd25519Address",
	"#FnUtilsEd25519Valid",
	"#FnUtilsHashBlake2b",
	"$FnUtilsHashName",
	"#FnUtilsHashSha3",
}

// WasmContextSandbox is the host side of the WasmLib Sandbox interface
// It acts as a change-resistant layer to wrap changes to the ISCP sandbox,
// to limit bothering users of WasmLib as little as possible with those changes.
type WasmContextSandbox struct {
	common  isc.SandboxBase
	ctx     isc.Sandbox
	ctxView isc.SandboxView
	cvt     WasmConvertor
	wc      *WasmContext
}

var _ ISandbox = new(WasmContextSandbox)

var EventSubscribers []func(msg string)

func NewWasmContextSandbox(wc *WasmContext, ctx interface{}) *WasmContextSandbox {
	s := &WasmContextSandbox{wc: wc}
	switch tctx := ctx.(type) {
	case isc.Sandbox:
		s.common = tctx
		s.ctx = tctx
	case isc.SandboxView:
		s.common = tctx
		s.ctxView = tctx
	default:
		panic(isc.ErrWrongTypeEntryPoint)
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

func (s *WasmContextSandbox) makeRequest(args []byte) isc.RequestParameters {
	req := wasmrequests.NewPostRequestFromBytes(args)
	chainID := s.cvt.IscpChainID(&req.ChainID)
	contract := s.cvt.IscpHname(req.Contract)
	function := s.cvt.IscpHname(req.Function)
	params, err := dict.FromBytes(req.Params)
	s.checkErr(err)

	allowance := s.cvt.IscpAllowance(wasmlib.NewScAssets(req.Allowance))
	transfer := s.cvt.IscpAllowance(wasmlib.NewScAssets(req.Transfer))
	if allowance.IsEmpty() {
		allowance = transfer
	}
	// Force a minimum transfer of 1million base tokens for dust and some gas
	// excess can always be reclaimed from the chain account by the user
	if !transfer.IsEmpty() && transfer.Assets.BaseTokens < 1*isc.Mi {
		transfer = transfer.Clone()
		transfer.Assets.BaseTokens = 1 * isc.Mi
	}

	s.Tracef("POST %s.%s, chain %s", contract.String(), function.String(), chainID.String())
	sendReq := isc.RequestParameters{
		AdjustToMinimumDustDeposit: true,
		TargetAddress:              chainID.AsAddress(),
		FungibleTokens:             transfer.Assets,
		Metadata: &isc.SendMetadata{
			TargetContract: contract,
			EntryPoint:     function,
			Params:         params,
			Allowance:      allowance,
			GasBudget:      gas.MaxGasPerCall,
		},
	}
	if req.Delay != 0 {
		timeLock := s.ctx.Timestamp()
		timeLock = timeLock.Add(time.Duration(req.Delay) * time.Second)
		sendReq.Options.Timelock = timeLock
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

func (s *WasmContextSandbox) fnAccountID(_ []byte) []byte {
	return s.cvt.ScAgentID(s.common.AccountID()).Bytes()
}

func (s *WasmContextSandbox) fnAllowance(_ []byte) []byte {
	allowance := s.ctx.AllowanceAvailable()
	return s.cvt.ScBalances(allowance).Bytes()
}

func (s *WasmContextSandbox) fnBalance(args []byte) []byte {
	if len(args) == 0 {
		return codec.EncodeUint64(s.common.BalanceBaseTokens())
	}
	tokenID := wasmtypes.TokenIDFromBytes(args)
	token := s.cvt.IscpTokenID(&tokenID)
	return codec.EncodeUint64(s.common.BalanceNativeToken(token).Uint64())
}

func (s *WasmContextSandbox) fnBalances(_ []byte) []byte {
	allowance := &isc.Allowance{}
	allowance.Assets = s.common.BalanceFungibleTokens()
	allowance.NFTs = s.common.OwnedNFTs()
	return s.cvt.ScBalances(allowance).Bytes()
}

func (s *WasmContextSandbox) fnBlockContext(_ []byte) []byte {
	panic("implement me")
}

func (s *WasmContextSandbox) fnCall(args []byte) []byte {
	req := wasmrequests.NewCallRequestFromBytes(args)
	contract := s.cvt.IscpHname(req.Contract)
	function := s.cvt.IscpHname(req.Function)
	params, err := dict.FromBytes(req.Params)
	s.checkErr(err)
	allowance := s.cvt.IscpAllowance(wasmlib.NewScAssets(req.Allowance))
	s.Tracef("CALL %s.%s", contract.String(), function.String())
	results := s.callUnlocked(contract, function, params, allowance)
	return results.Bytes()
}

func (s *WasmContextSandbox) callUnlocked(contract, function isc.Hname, params dict.Dict, transfer *isc.Allowance) dict.Dict {
	// TODO is this really necessary? We should not be able to call in parallel
	s.wc.proc.instanceLock.Unlock()
	defer s.wc.proc.instanceLock.Lock()

	if s.ctx != nil {
		return s.ctx.Call(contract, function, params, transfer)
	}
	return s.ctxView.CallView(contract, function, params)
}

func (s *WasmContextSandbox) fnCaller(_ []byte) []byte {
	return s.cvt.ScAgentID(s.ctx.Caller()).Bytes()
}

func (s *WasmContextSandbox) fnChainID(_ []byte) []byte {
	return s.cvt.ScChainID(s.common.ChainID()).Bytes()
}

func (s *WasmContextSandbox) fnChainOwnerID(_ []byte) []byte {
	return s.cvt.ScAgentID(s.common.ChainOwnerID()).Bytes()
}

func (s *WasmContextSandbox) fnContract(_ []byte) []byte {
	return s.cvt.ScHname(s.common.Contract()).Bytes()
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
	// TODO is this really necessary? We should not be able to call in parallel
	s.wc.proc.instanceLock.Unlock()
	defer s.wc.proc.instanceLock.Lock()

	s.ctx.DeployContract(programHash, name, description, params)
}

func (s *WasmContextSandbox) fnEntropy(_ []byte) []byte {
	return s.cvt.ScHash(s.ctx.GetEntropy()).Bytes()
}

func (s *WasmContextSandbox) fnEstimateDust(args []byte) []byte {
	dust := s.ctx.EstimateRequiredDustDeposit(s.makeRequest(args))
	return codec.EncodeUint64(dust)
}

func (s *WasmContextSandbox) fnEvent(args []byte) []byte {
	msg := string(args)
	s.ctx.Event(msg)
	for _, eventSubscribers := range EventSubscribers {
		eventSubscribers(msg)
	}
	return nil
}

func (s *WasmContextSandbox) fnLog(args []byte) []byte {
	s.common.Log().Infof(string(args))
	return nil
}

func (s *WasmContextSandbox) fnMinted(_ []byte) []byte {
	panic("fixme: wc.fnMinted")
	// return s.ctx.Minted().Bytes()
}

func (s *WasmContextSandbox) fnPanic(args []byte) []byte {
	s.common.Log().Panicf("WASM: panic in VM: %s", string(args))
	return nil
}

func (s *WasmContextSandbox) fnParams(_ []byte) []byte {
	return s.common.Params().Dict.Bytes()
}

func (s *WasmContextSandbox) fnPost(args []byte) []byte {
	s.ctx.Send(s.makeRequest(args))
	return nil
}

func (s *WasmContextSandbox) fnRequest(_ []byte) []byte {
	panic("fixme: wc.fnRequest")
	// return s.ctx.Request().Bytes()
}

func (s *WasmContextSandbox) fnRequestID(_ []byte) []byte {
	return s.cvt.ScRequestID(s.ctx.Request().ID()).Bytes()
}

func (s *WasmContextSandbox) fnRequestSender(_ []byte) []byte {
	return s.cvt.ScAgentID(s.ctx.Request().SenderAccount()).Bytes()
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
		metadata := isc.RequestParameters{
			AdjustToMinimumDustDeposit: true,
			TargetAddress:              address,
			FungibleTokens:             allowance.Assets,
		}
		if len(allowance.NFTs) == 0 {
			s.ctx.Send(metadata)
			return nil
		}
		s.ctx.SendAsNFT(metadata, allowance.NFTs[0])
	}
	return nil
}

func (s *WasmContextSandbox) fnStateAnchor(_ []byte) []byte {
	panic("implement me")
}

func (s *WasmContextSandbox) fnTimestamp(_ []byte) []byte {
	return codec.EncodeUint64(uint64(s.common.Timestamp().UnixNano()))
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

func (s WasmContextSandbox) fnUtilsBech32Decode(args []byte) []byte {
	hrp, addr, err := iotago.ParseBech32(string(args))
	s.checkErr(err)
	if hrp != parameters.L1.Protocol.Bech32HRP {
		s.Panicf("Invalid protocol prefix: %s", string(hrp))
	}
	return s.cvt.ScAddress(addr).Bytes()
}

func (s WasmContextSandbox) fnUtilsBech32Encode(args []byte) []byte {
	scAddress := wasmtypes.AddressFromBytes(args)
	addr := s.cvt.IscpAddress(&scAddress)
	return []byte(addr.Bech32(parameters.L1.Protocol.Bech32HRP))
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
