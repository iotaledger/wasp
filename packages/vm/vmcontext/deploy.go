package vmcontext

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
)

func (vmctx *VMContext) DeployContract(vmtype string, programBinary []byte, name string, description string, initParams codec.ImmutableCodec) error {
	vmctx.log.Debugf("vmcontext.DeployContract")

	if vmctx.CurrentContractHname() == root.Hname {
		// from root contract calling VMContext directly
		deploymentHash, err := vmctx.processors.NewProcessor(programBinary, vmtype)
		if err != nil {
			return err
		}
		// storing contract in the registry
		err = root.StoreContract(codec.NewMustCodec(vmctx), &root.ContractRecord{
			VMType:         vmtype,
			DeploymentHash: deploymentHash,
			Description:    description,
			Name:           name,
		}, programBinary)
		if err != nil {
			return err
		}
		// calling constructor
		_, err = vmctx.CallContract(coretypes.Hn(name), coretypes.EntryPointInit, initParams, nil)
		if err != nil {
			vmctx.log.Warnf("sandbox.DeployContract. Error while calling init function: %v", err)
		}
		// ignoring error because init method may not exist
		return nil
	}
	// calling root contract from another contract to install contract
	// adding parameters specific to deployment
	par := codec.NewCodec(dict.New())
	err := initParams.Iterate("", func(key kv.Key, value []byte) bool {
		par.Set(key, value)
		return true
	})
	if err != nil {
		return err
	}
	par.SetString(root.ParamVMType, vmtype)
	par.Set(root.ParamProgramBinary, programBinary)
	par.SetString(root.ParamName, name)
	par.SetString(root.ParamDescription, description)
	_, err = vmctx.CallContract(root.Hname, coretypes.Hn(root.FuncDeployContract), par, nil)
	return err
}
