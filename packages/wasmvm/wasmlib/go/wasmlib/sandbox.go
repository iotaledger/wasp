// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmlib

import (
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmrequests"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmlib/go/wasmlib/wasmtypes"
)

const (
	FnAccountID           = int32(-1)
	FnAllowance           = int32(-2)
	FnBalance             = int32(-3)
	FnBalances            = int32(-4)
	FnBlockContext        = int32(-5)
	FnCall                = int32(-6)
	FnCaller              = int32(-7)
	FnChainID             = int32(-8)
	FnChainOwnerID        = int32(-9)
	FnContract            = int32(-10)
	FnDeployContract      = int32(-11)
	FnEntropy             = int32(-12)
	FnEstimateDust        = int32(-13)
	FnEvent               = int32(-14)
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
	FnTransferAllowed     = int32(-27)
	FnUtilsBech32Decode   = int32(-28)
	FnUtilsBech32Encode   = int32(-29)
	FnUtilsBlsAddress     = int32(-30)
	FnUtilsBlsAggregate   = int32(-31)
	FnUtilsBlsValid       = int32(-32)
	FnUtilsEd25519Address = int32(-33)
	FnUtilsEd25519Valid   = int32(-34)
	FnUtilsHashBlake2b    = int32(-35)
	FnUtilsHashName       = int32(-36)
	FnUtilsHashSha3       = int32(-37)
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

func (s ScSandbox) Balance(tokenID wasmtypes.ScTokenID) uint64 {
	return wasmtypes.Uint64FromBytes(Sandbox(FnBalance, tokenID.Bytes()))
}

// access the current balances for all assets
func (s ScSandbox) Balances() *ScBalances {
	balances := NewScAssets(Sandbox(FnBalances, nil)).Balances()
	return &balances
}

// calls a smart contract function
func (s ScSandbox) callWithAllowance(hContract, hFunction wasmtypes.ScHname, params *ScDict, allowance *ScTransfer) *ScImmutableDict {
	req := &wasmrequests.CallRequest{
		Contract:  hContract,
		Function:  hFunction,
		Params:    params.Bytes(),
		Allowance: allowance.Bytes(),
	}
	res := Sandbox(FnCall, req.Bytes())
	return NewScDictFromBytes(res).Immutable()
}

// retrieve the agent id of the owner of the chain this contract lives on
func (s ScSandbox) ChainOwnerID() wasmtypes.ScAgentID {
	return wasmtypes.AgentIDFromBytes(Sandbox(FnChainOwnerID, nil))
}

// retrieve the hname of this contract
func (s ScSandbox) Contract() wasmtypes.ScHname {
	return wasmtypes.HnameFromBytes(Sandbox(FnContract, nil))
}

// retrieve the chain id of the chain this contract lives on
func (s ScSandbox) CurrentChainID() wasmtypes.ScChainID {
	return wasmtypes.ChainIDFromBytes(Sandbox(FnChainID, nil))
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
	Sandbox(FnResults, results.Bytes())
}

// deterministic timestamp in nanosecond fixed at the moment of calling the smart contract
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
	return s.callWithAllowance(hContract, hFunction, params, nil)
}

type ScSandboxFunc struct {
	ScSandbox
}

// access the allowance assets
func (s ScSandboxFunc) Allowance() *ScBalances {
	buf := Sandbox(FnAllowance, nil)
	balances := NewScAssets(buf).Balances()
	return &balances
}

//func (s ScSandbox) BlockContext(construct func(sandbox ScSandbox) interface{}, onClose func(interface{})) interface{} {
//	panic("implement me")
//}

// calls a smart contract function
func (s ScSandboxFunc) Call(hContract, hFunction wasmtypes.ScHname, params *ScDict, allowance *ScTransfer) *ScImmutableDict {
	return s.callWithAllowance(hContract, hFunction, params, allowance)
}

// retrieve the agent id of the caller of the smart contract
func (s ScSandboxFunc) Caller() wasmtypes.ScAgentID {
	return wasmtypes.AgentIDFromBytes(Sandbox(FnCaller, nil))
}

// deploys a smart contract
func (s ScSandboxFunc) DeployContract(programHash wasmtypes.ScHash, name, description string, initParams *ScDict) {
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

func (s ScSandboxFunc) EstimateDust(fn *ScFunc) uint64 {
	req := &wasmrequests.PostRequest{
		Contract:  fn.hContract,
		Function:  fn.hFunction,
		Params:    fn.params.Bytes(),
		Allowance: fn.allowance.Bytes(),
		Transfer:  fn.transfer.Bytes(),
		Delay:     fn.delay,
	}
	return wasmtypes.Uint64FromBytes(Sandbox(FnEstimateDust, req.Bytes()))
}

// signals an event on the node that external entities can subscribe to
func (s ScSandboxFunc) Event(msg string) {
	Sandbox(FnEvent, []byte(msg))
}

// retrieve the assets that were minted in this transaction
func (s ScSandboxFunc) Minted() ScBalances {
	return NewScAssets(Sandbox(FnMinted, nil)).Balances()
}

// Post (delayed) posts a SC function request
func (s ScSandboxFunc) Post(chainID wasmtypes.ScChainID, hContract, hFunction wasmtypes.ScHname, params *ScDict, allowance, transfer ScTransfer, delay uint32) {
	req := &wasmrequests.PostRequest{
		ChainID:   chainID,
		Contract:  hContract,
		Function:  hFunction,
		Params:    params.Bytes(),
		Allowance: allowance.Bytes(),
		Transfer:  transfer.Bytes(),
		Delay:     delay,
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

// Send transfers SC assets to the specified address
func (s ScSandboxFunc) Send(address wasmtypes.ScAddress, transfer *ScTransfer) {
	// we need some assets to send
	if transfer.IsEmpty() {
		return
	}

	req := wasmrequests.SendRequest{
		Address:  address,
		Transfer: transfer.Bytes(),
	}
	Sandbox(FnSend, req.Bytes())
}

//func (s ScSandboxFunc) StateAnchor() interface{} {
//	panic("implement me")
//}

// TransferAllowed transfers allowed assets from caller to the specified account
func (s ScSandboxFunc) TransferAllowed(agentID wasmtypes.ScAgentID, transfer *ScTransfer, create bool) {
	// we need some assets to send
	if transfer.IsEmpty() {
		return
	}

	req := wasmrequests.TransferRequest{
		AgentID:  agentID,
		Create:   create,
		Transfer: transfer.Bytes(),
	}
	Sandbox(FnTransferAllowed, req.Bytes())
}
