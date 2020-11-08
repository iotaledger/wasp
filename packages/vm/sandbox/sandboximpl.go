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

func (s *sandbox) GetContractID() coretypes.ContractID {
	return coretypes.NewContractID(s.vmctx.ChainID(), s.vmctx.ContractIndex())
}

func (s *sandbox) GetChainID() coretypes.ChainID {
	return s.vmctx.ChainID()
}

func (s *sandbox) GetContractIndex() uint16 {
	return s.vmctx.ContractIndex()
}

func (s *sandbox) GetOwnerAddress() *address.Address {
	return s.vmctx.OwnerAddress()
}

func (s *sandbox) GetTimestamp() int64 {
	return s.vmctx.Timestamp()
}

func (s *sandbox) GetEntropy() hashing.HashValue {
	return s.vmctx.Entropy()
}

func (s *sandbox) DumpAccount() string {
	return s.vmctx.DumpAccount()
}

// request context

func (s *sandbox) AccessRequest() vmtypes.RequestAccess {
	return s
}

func (s *sandbox) AccessState() codec.MutableMustCodec {
	return codec.NewMustCodec(s)
}

func (s *sandbox) AccessSCAccount() vmtypes.AccountAccess {
	return s.vmctx
}

func (s *sandbox) Accounts() vmtypes.Accounts {
	return s.vmctx.Accounts()
}

func (s *sandbox) SendRequest(par vmtypes.NewRequestParams) bool {
	return s.vmctx.SendRequest(par)
}

func (s *sandbox) SendRequestToSelf(reqCode coretypes.EntryPointCode, args dict.Dict) bool {
	return s.vmctx.SendRequestToSelf(reqCode, args)
}

func (s *sandbox) SendRequestToSelfWithDelay(entryPoint coretypes.EntryPointCode, args dict.Dict, delaySec uint32) bool {
	return s.vmctx.SendRequestToSelfWithDelay(entryPoint, args, delaySec)
}

func (s *sandbox) Event(msg string) {
	s.vmctx.Log().Infof("VMMSG contract #%d '%s'", s.GetContractIndex(), msg)
	s.vmctx.Publish(msg)
}

func (s *sandbox) Eventf(format string, args ...interface{}) {
	s.vmctx.Log().Infof("VMMSG: "+format, args...)
	s.vmctx.Publishf(format, args...)
}
