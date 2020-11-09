package sandbox

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
)

// DeployContract
// - if called from the root contract, call VMContext
// - if called from other than root contract, it redirects call to the root contract
// - call "init" endpoint (constructor) with provided parameters
func (s *sandbox) DeployContract(vmtype string, programBinary []byte, name string, description string, initParams codec.ImmutableCodec) (uint16, error) {
	s.vmctx.Log().Debugf("sandbox.DeployBuiltinContract")

	var ret uint16
	if s.GetContractIndex() == 0 {
		// from root contract calling VMContext directly
		var err error
		if ret, err = s.vmctx.InstallContract(vmtype, programBinary, name, description); err != nil {
			return 0, err
		}
	} else {
		// calling root contract from another contract to install contract
		par := codec.NewCodec(dict.NewDict())
		par.SetString(root.ParamVMType, vmtype)
		par.Set(root.ParamProgramBinary, programBinary)
		par.SetString(root.ParamName, name)
		par.SetString(root.ParamDescription, description)

		resp, err := s.CallContract(0, root.EntryPointDeployContract, par, nil)
		if err != nil {
			return 0, err
		}
		t, ok, err := resp.GetInt64("index")
		if err != nil || !ok {
			s.Panic("internal error")
			return 0, nil
		}
		ret = uint16(t)
	}
	// calling constructor
	// error ignored, if for example init entry point does not exist
	_, _ = s.CallContract(ret, coretypes.EntryPointCodeInit, initParams, nil)

	return ret, nil
}

func (s *sandbox) CallContract(contractIndex uint16, entryPoint coretypes.Hname, params codec.ImmutableCodec, budget coretypes.ColoredBalancesSpendable) (codec.ImmutableCodec, error) {
	return s.vmctx.CallContract(contractIndex, entryPoint, params, budget)
}
