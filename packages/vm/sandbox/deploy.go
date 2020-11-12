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
func (s *sandbox) DeployContract(vmtype string, programBinary []byte, name string, description string, initParams codec.ImmutableCodec) error {
	s.vmctx.Log().Debugf("sandbox.DeployContract")

	if s.GetContractHname() == root.Hname {
		// from root contract calling VMContext directly
		var err error
		if err = s.vmctx.InstallContract(vmtype, programBinary, name, description); err != nil {
			return err
		}
	} else {
		// calling root contract from another contract to install contract
		par := codec.NewCodec(dict.New())
		par.SetString(root.ParamVMType, vmtype)
		par.Set(root.ParamProgramBinary, programBinary)
		par.SetString(root.ParamName, name)
		par.SetString(root.ParamDescription, description)

		_, err := s.Call(root.Hname, root.EntryPointDeployContract, par, nil)
		if err != nil {
			return err
		}
	}
	// calling constructor
	// error ignored, if for example init entry point does not exist
	_, _ = s.Call(coretypes.Hn(name), coretypes.EntryPointCodeInit, initParams, nil)

	return nil
}

func (s *sandbox) Call(contractHname coretypes.Hname, entryPoint coretypes.Hname, params codec.ImmutableCodec, budget coretypes.ColoredBalancesSpendable) (codec.ImmutableCodec, error) {
	return s.vmctx.CallContract(contractHname, entryPoint, params, budget)
}
