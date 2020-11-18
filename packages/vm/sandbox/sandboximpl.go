package sandbox

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/vmcontext"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

type sandbox struct {
	vmctx *vmcontext.VMContext
}

func init() {
	vmcontext.NewSandbox = new
}

func new(vmctx *vmcontext.VMContext) vmtypes.Sandbox {
	return &sandbox{
		vmctx: vmctx,
	}
}

// Sandbox interface

func (s *sandbox) Panic(v interface{}) {
	panic(v)
}

func (s *sandbox) Rollback() {
	s.vmctx.Rollback()
}

func (s *sandbox) MyContractID() coretypes.ContractID {
	return s.vmctx.CurrentContractID()
}

func (s *sandbox) Caller() coretypes.AgentID {
	return s.vmctx.Caller()
}

func (s *sandbox) MyAgentID() coretypes.AgentID {
	return coretypes.NewAgentIDFromContractID(s.vmctx.CurrentContractID())
}

func (s *sandbox) IsRequestContext() bool {
	return s.vmctx.IsRequestContext()
}

func (s *sandbox) ChainID() coretypes.ChainID {
	return s.vmctx.ChainID()
}

func (s *sandbox) ChainOwnerID() coretypes.AgentID {
	return s.vmctx.ChainOwnerID()
}

func (s *sandbox) GetTimestamp() int64 {
	return s.vmctx.Timestamp()
}

func (s *sandbox) GetEntropy() hashing.HashValue {
	return s.vmctx.Entropy()
}

// request context

func (s *sandbox) AccessRequest() vmtypes.RequestAccess {
	return s
}

func (s *sandbox) AccessState() codec.MutableMustCodec {
	return codec.NewMustCodec(s)
}

func (s *sandbox) Accounts() vmtypes.Accounts {
	return s.vmctx.Accounts()
}

func (s *sandbox) TransferToAddress(targetAddr address.Address, transfer coretypes.ColoredBalances) bool {
	return s.vmctx.TransferToAddress(targetAddr, transfer)
}

func (s *sandbox) PostRequest(par vmtypes.NewRequestParams) bool {
	return s.vmctx.PostRequest(par)
}

func (s *sandbox) PostRequestToSelf(reqCode coretypes.Hname, args dict.Dict) bool {
	return s.vmctx.PostRequestToSelf(reqCode, args)
}

func (s *sandbox) PostRequestToSelfWithDelay(entryPoint coretypes.Hname, args dict.Dict, delaySec uint32) bool {
	return s.vmctx.PostRequestToSelfWithDelay(entryPoint, args, delaySec)
}

func (s *sandbox) Event(msg string) {
	s.vmctx.Log().Infof("VMMSG contract %s '%s'", s.MyContractID().String(), msg)
	s.vmctx.Publish(msg)
}

func (s *sandbox) Eventf(format string, args ...interface{}) {
	s.vmctx.Log().Infof("VMMSG: "+format, args...)
	s.vmctx.Publishf(format, args...)
}
