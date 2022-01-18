// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package wasmlib

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

type ScSandbox struct{}

func (s ScSandbox) AccountID() ScAgentID {
	return NewScAgentIDFromBytes(Sandbox(FnAccountID, nil))
}

func (s ScSandbox) Balance(color ScColor) uint64 {
	bal, _ := ExtractUint64(Sandbox(FnBalance, color.Bytes()))
	return bal
}

func (s ScSandbox) Balances() ScAssets {
	return NewScAssetsFromBytes(Sandbox(FnBalances, nil))
}

func (s ScSandbox) BlockContext(construct func(sandbox ScSandbox) interface{}, onClose func(interface{})) interface{} {
	panic("implement me")
}

func (s ScSandbox) Call(contract, function ScHname, params ScDict, transfer ScAssets) ScDict {
	enc := NewBytesEncoder()
	enc.Hname(contract)
	enc.Hname(function)
	enc.Bytes(params.Bytes())
	enc.Bytes(transfer.Bytes())
	return NewScDictFromBytes(Sandbox(FnCall, enc.Data()))
}

func (s ScSandbox) Caller() ScAgentID {
	return NewScAgentIDFromBytes(Sandbox(FnCaller, nil))
}

func (s ScSandbox) ChainID() ScChainID {
	return NewScChainIDFromBytes(Sandbox(FnChainID, nil))
}

func (s ScSandbox) ChainOwnerID() ScAgentID {
	return NewScAgentIDFromBytes(Sandbox(FnChainOwnerID, nil))
}

func (s ScSandbox) Contract() ScHname {
	return NewScHnameFromBytes(Sandbox(FnContract, nil))
}

func (s ScSandbox) ContractCreator() ScAgentID {
	return NewScAgentIDFromBytes(Sandbox(FnContractCreator, nil))
}

func (s ScSandbox) DeployContract(programHash ScHash, name, description string, initParams ScDict) {
	enc := NewBytesEncoder()
	enc.Hash(programHash)
	enc.String(name)
	enc.String(description)
	enc.Bytes(initParams.Bytes())
	Sandbox(FnDeployContract, enc.Data())
}

func (s ScSandbox) Entropy() ScHash {
	return NewScHashFromBytes(Sandbox(FnEntropy, nil))
}

func (s ScSandbox) Event(msg string) {
	Sandbox(FnEvent, []byte(msg))
}

func (s ScSandbox) IncomingTransfer() ScAssets {
	return NewScAssetsFromBytes(Sandbox(FnIncomingTransfer, nil))
}

func (s ScSandbox) Log(text string) {
	Sandbox(FnLog, []byte(text))
}

func (s ScSandbox) Minted() ScAssets {
	return NewScAssetsFromBytes(Sandbox(FnMinted, nil))
}

func (s ScSandbox) Panic(text string) {
	Sandbox(FnPanic, []byte(text))
}

func (s ScSandbox) Params() ScDict {
	return NewScDictFromBytes(Sandbox(FnParams, nil))
}

func (s ScSandbox) Post(chainID ScChainID, contract, function ScHname, params ScDict, transfer ScAssets, delay uint32) {
	enc := NewBytesEncoder()
	enc.ChainID(chainID)
	enc.Hname(contract)
	enc.Hname(function)
	enc.Bytes(params.Bytes())
	enc.Bytes(transfer.Bytes())
	enc.Uint32(delay)
	Sandbox(FnSend, enc.Data())
}

//func (s ScSandbox) Request() ScRequest {
//	panic("implement me")
//}

func (s ScSandbox) RequestID() ScRequestID {
	return NewScRequestIDFromBytes(Sandbox(FnRequestID, nil))
}

func (s ScSandbox) Send(target ScAddress, tokens ScAssets) {
	enc := NewBytesEncoder()
	enc.Address(target)
	enc.Bytes(tokens.Bytes())
	Sandbox(FnSend, enc.Data())
}

func (s ScSandbox) StateAnchor() interface{} {
	panic("implement me")
}

func (s ScSandbox) Timestamp() int64 {
	ts, _ := ExtractInt64(Sandbox(FnTimestamp, nil))
	return ts
}

func (s ScSandbox) Trace(text string) {
	Sandbox(FnTrace, []byte(text))
}
