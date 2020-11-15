package sandbox

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/accountsc"
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
	return coretypes.NewAgentIDFromContractID(coretypes.NewContractID(s.vmctx.ChainID(), accountsc.Hname))
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

func (s *sandbox) PostRequest(par vmtypes.NewRequestParams) bool {
	return s.vmctx.PostRequest(par)
}

func (s *sandbox) PostRequestToSelf(reqCode coretypes.Hname, args dict.Dict) bool {
	return s.vmctx.SendRequestToSelf(reqCode, args)
}

func (s *sandbox) PostRequestToSelfWithDelay(entryPoint coretypes.Hname, args dict.Dict, delaySec uint32) bool {
	return s.vmctx.SendRequestToSelfWithDelay(entryPoint, args, delaySec)
}

func (s *sandbox) Event(msg string) {
	s.vmctx.Log().Infof("VMMSG contract %s '%s'", s.MyContractID().String(), msg)
	s.vmctx.Publish(msg)
}

func (s *sandbox) Eventf(format string, args ...interface{}) {
	s.vmctx.Log().Infof("VMMSG: "+format, args...)
	s.vmctx.Publishf(format, args...)
}
