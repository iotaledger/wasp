package vmcontext

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
)

// CreateContract deploys contract by its program hash
func (vmctx *VMContext) CreateContract(programHash hashing.HashValue, name string, description string, initParams dict.Dict) error {
	vmtype, programBinary, err := vmctx.getBinary(programHash)
	if err != nil {
		return err
	}
	if vmctx.CurrentContractHname() == root.Interface.Hname() {
		// from root contract only loading VM
		vmctx.log.Debugf("vmcontext.DeployContract: %s from root", programHash.String())
		return vmctx.processors.NewProcessor(programHash, programBinary, vmtype)
	}
	vmctx.log.Debugf("vmcontext.DeployContract: %s, name: %s, dscr: '%s'", programHash.String(), name, description)

	// calling root contract from another contract to install contract
	// adding parameters specific to deployment
	par := initParams.Clone()
	if err != nil {
		return err
	}
	par.Set(root.ParamProgramHash, codec.EncodeHashValue(&programHash))
	par.Set(root.ParamName, codec.EncodeString(name))
	par.Set(root.ParamDescription, codec.EncodeString(description))
	_, err = vmctx.Call(root.Interface.Hname(), coretypes.Hn(root.FuncDeployContract), par, nil)
	return err

}
