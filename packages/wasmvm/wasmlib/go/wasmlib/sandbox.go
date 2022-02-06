// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmlib

import (
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmrequests"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
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
	FnResults             = int32(-22)
	FnSend                = int32(-23)
	FnStateAnchor         = int32(-24)
	FnTimestamp           = int32(-25)
	FnTrace               = int32(-26)
	FnUtilsBase58Decode   = int32(-27)
	FnUtilsBase58Encode   = int32(-28)
	FnUtilsBlsAddress     = int32(-29)
	FnUtilsBlsAggregate   = int32(-30)
	FnUtilsBlsValid       = int32(-31)
	FnUtilsEd25519Address = int32(-32)
	FnUtilsEd25519Valid   = int32(-33)
	FnUtilsHashBlake2b    = int32(-34)
	FnUtilsHashName       = int32(-35)
	FnUtilsHashSha3       = int32(-36)
)

type ScSandbox struct{}

// TODO go over core contract schemas to set correct unsigned types

func Log(text string) {
	Sandbox(FnLog, []byte(text))
}

func Panic(text string) {
	Sandbox(FnPanic, []byte(text))
}

func Trace(text string) {
	Sandbox(FnTrace, []byte(text))
}

func NewParamsProxy() wasmtypes.Proxy {
	return wasmtypes.NewProxy(NewScDictFromBytes(Sandbox(FnParams, nil)))
}

// retrieve the agent id of this contract account
func (s ScSandbox) AccountID() wasmtypes.ScAgentID {
	return wasmtypes.AgentIDFromBytes(Sandbox(FnAccountID, nil))
}

func (s ScSandbox) Balance(color wasmtypes.ScColor) uint64 {
	return wasmtypes.Uint64FromBytes(Sandbox(FnBalance, color.Bytes()))
}

// access the current balances for all assets
func (s ScSandbox) Balances() ScBalances {
	return NewScAssetsFromBytes(Sandbox(FnBalances, nil)).Balances()
}

// calls a smart contract function
func (s ScSandbox) call(hContract, hFunction wasmtypes.ScHname, params *ScDict, transfer ScTransfers) *ScImmutableDict {
	if params == nil {
		params = NewScDict()
	}
	req := &wasmrequests.CallRequest{
		Contract: hContract,
		Function: hFunction,
		Params:   params.Bytes(),
		Transfer: ScAssets(transfer).Bytes(),
	}
	res := Sandbox(FnCall, req.Bytes())
	return NewScDictFromBytes(res).Immutable()
}

// retrieve the chain id of the chain this contract lives on
func (s ScSandbox) ChainID() wasmtypes.ScChainID {
	return wasmtypes.ChainIDFromBytes(Sandbox(FnChainID, nil))
}

// retrieve the agent id of the owner of the chain this contract lives on
func (s ScSandbox) ChainOwnerID() wasmtypes.ScAgentID {
	return wasmtypes.AgentIDFromBytes(Sandbox(FnChainOwnerID, nil))
}

// retrieve the hname of this contract
func (s ScSandbox) Contract() wasmtypes.ScHname {
	return wasmtypes.HnameFromBytes(Sandbox(FnContract, nil))
}

// retrieve the agent id of the creator of this contract
func (s ScSandbox) ContractCreator() wasmtypes.ScAgentID {
	return wasmtypes.AgentIDFromBytes(Sandbox(FnContractCreator, nil))
}

// logs informational text message
func (s ScSandbox) Log(text string) {
	Sandbox(FnLog, []byte(text))
}

// logs error text message and then panics
func (s ScSandbox) Panic(text string) {
	Sandbox(FnPanic, []byte(text))
}

// retrieve parameters passed to the smart contract function that was called
func (s ScSandbox) Params() *ScImmutableDict {
	return NewScDictFromBytes(Sandbox(FnParams, nil)).Immutable()
}

func (s ScSandbox) RawState() ScImmutableState {
	return ScImmutableState{}
}

// panics if condition is not satisfied
func (s ScSandbox) Require(cond bool, msg string) {
	if !cond {
		s.Panic(msg)
	}
}

func (s ScSandbox) Results(results *ScDict) {
	if results == nil {
		results = NewScDict()
	}
	Sandbox(FnResults, results.Bytes())
}

// deterministic time stamp fixed at the moment of calling the smart contract
func (s ScSandbox) Timestamp() uint64 {
	return wasmtypes.Uint64FromBytes(Sandbox(FnTimestamp, nil))
}

// logs debugging trace text message
func (s ScSandbox) Trace(text string) {
	Sandbox(FnTrace, []byte(text))
}

// access diverse utility functions
func (s ScSandbox) Utility() ScSandboxUtils {
	return ScSandboxUtils{}
}

type ScSandboxView struct {
	ScSandbox
}

// calls a smart contract view
func (s ScSandboxView) Call(hContract, hFunction wasmtypes.ScHname, params *ScDict) *ScImmutableDict {
	return s.call(hContract, hFunction, params, nil)
}

type ScSandboxFunc struct {
	ScSandbox
}

//func (s ScSandbox) BlockContext(construct func(sandbox ScSandbox) interface{}, onClose func(interface{})) interface{} {
//	panic("implement me")
//}

// calls a smart contract function
func (s ScSandboxFunc) Call(hContract, hFunction wasmtypes.ScHname, params *ScDict, transfer ScTransfers) *ScImmutableDict {
	return s.call(hContract, hFunction, params, transfer)
}

// retrieve the agent id of the caller of the smart contract
func (s ScSandboxFunc) Caller() wasmtypes.ScAgentID {
	return wasmtypes.AgentIDFromBytes(Sandbox(FnCaller, nil))
}

// deploys a smart contract
func (s ScSandboxFunc) DeployContract(programHash wasmtypes.ScHash, name, description string, initParams *ScDict) {
	if initParams == nil {
		initParams = NewScDict()
	}
	req := &wasmrequests.DeployRequest{
		ProgHash:    programHash,
		Name:        name,
		Description: description,
		Params:      initParams.Bytes(),
	}
	Sandbox(FnDeployContract, req.Bytes())
}

// returns random entropy data for current request.
func (s ScSandboxFunc) Entropy() wasmtypes.ScHash {
	return wasmtypes.HashFromBytes(Sandbox(FnEntropy, nil))
}

// signals an event on the node that external entities can subscribe to
func (s ScSandboxFunc) Event(msg string) {
	Sandbox(FnEvent, []byte(msg))
}

// access the incoming balances for all assets
func (s ScSandboxFunc) IncomingTransfer() ScBalances {
	buf := Sandbox(FnIncomingTransfer, nil)
	return NewScAssetsFromBytes(buf).Balances()
}

// retrieve the assets that were minted in this transaction
func (s ScSandboxFunc) Minted() ScBalances {
	return NewScAssetsFromBytes(Sandbox(FnMinted, nil)).Balances()
}

// (delayed) posts a smart contract function request
func (s ScSandboxFunc) Post(chainID wasmtypes.ScChainID, hContract, hFunction wasmtypes.ScHname, params *ScDict, transfer ScTransfers, delay uint32) {
	if params == nil {
		params = NewScDict()
	}
	if len(transfer) == 0 {
		s.Panic("missing transfer")
	}
	req := &wasmrequests.PostRequest{
		ChainID:  chainID,
		Contract: hContract,
		Function: hFunction,
		Params:   params.Bytes(),
		Transfer: ScAssets(transfer).Bytes(),
		Delay:    delay,
	}
	Sandbox(FnPost, req.Bytes())
}

var (
	entropy []byte
	offset  = 0
)

// generates a random value from 0 to max (exclusive max) using a deterministic RNG
func (s ScSandboxFunc) Random(max uint64) (rnd uint64) {
	if max == 0 {
		s.Panic("random: max parameter should be non-zero")
	}

	// note that entropy gets reset for every request
	if len(entropy) == 0 {
		// first time in this request, initialize with current request entropy
		entropy = s.Entropy().Bytes()
		offset = 0
	}
	if offset == 32 {
		// ran out of entropy data, hash entropy for next pseudo-random entropy
		entropy = s.Utility().HashBlake2b(entropy).Bytes()
		offset = 0
	}
	rnd = wasmtypes.Uint64FromBytes(entropy[offset:offset+8]) % max
	offset += 8
	return rnd
}

func (s ScSandboxFunc) RawState() ScState {
	return ScState{}
}

//func (s ScSandboxFunc) Request() ScRequest {
//	panic("implement me")
//}

// retrieve the request id of this transaction
func (s ScSandboxFunc) RequestID() wasmtypes.ScRequestID {
	return wasmtypes.RequestIDFromBytes(Sandbox(FnRequestID, nil))
}

// transfer assetss to the specified Tangle ledger address
func (s ScSandboxFunc) Send(address wasmtypes.ScAddress, transfer ScTransfers) {
	// we need some assets to send
	assets := uint64(0)
	for _, amount := range transfer {
		assets += amount
	}
	if assets == 0 {
		// only try to send when non-zero assets
		return
	}

	req := wasmrequests.SendRequest{
		Address:  address,
		Transfer: ScAssets(transfer).Bytes(),
	}
	Sandbox(FnSend, req.Bytes())
}

//func (s ScSandboxFunc) StateAnchor() interface{} {
//	panic("implement me")
//}
