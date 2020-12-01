package sandbox

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

// DeployContract
// - locates program binary
// - if called from the root contract, call VMContext
// - if called from other than root contract, it redirects call to the root contract
// - call "init" endpoint (constructor) with provided parameters
func (s *sandbox) CreateContract(programHash hashing.HashValue, name string, description string, initParams codec.ImmutableCodec) error {
	return s.vmctx.CreateContract(programHash, name, description, initParams)
}

// Call calls an entry point of contact, passes parameters and funds
func (s *sandbox) Call(contractHname coretypes.Hname, entryPoint coretypes.Hname, params codec.ImmutableCodec, transfer coretypes.ColoredBalances) (codec.ImmutableCodec, error) {
	return s.vmctx.Call(contractHname, entryPoint, params, transfer)
}
