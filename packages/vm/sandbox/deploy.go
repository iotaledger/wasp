package sandbox

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

// DeployContract
// - if called from the root contract, call VMContext
// - if called from other than root contract, it redirects call to the root contract
// - call "init" endpoint (constructor) with provided parameters
func (s *sandbox) DeployContract(vmtype string, programBinary []byte, name string, description string, initParams codec.ImmutableCodec) error {
	return s.vmctx.DeployContract(vmtype, programBinary, name, description, initParams)
}

func (s *sandbox) Call(contractHname coretypes.Hname, entryPoint coretypes.Hname, params codec.ImmutableCodec, transfer coretypes.ColoredBalances) (codec.ImmutableCodec, error) {
	return s.vmctx.CallContract(contractHname, entryPoint, params, transfer)
}

func (s *sandbox) CallView(contractHname coretypes.Hname, entryPoint coretypes.Hname, params codec.ImmutableCodec) (codec.ImmutableCodec, error) {
	return s.vmctx.CallView(contractHname, entryPoint, params)
}
