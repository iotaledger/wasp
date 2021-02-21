package vmcontext

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

// DeployContract deploys contract by its program hash
// - if called from 'root' contract only loads VM from binary
// - otherwise calls 'root' contract 'DeployContract' entry point to do the job.
func (vmctx *VMContext) DeployContract(programHash hashing.HashValue, name string, description string, initParams dict.Dict) error {
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
	par.Set(root.ParamProgramHash, codec.EncodeHashValue(programHash))
	par.Set(root.ParamName, codec.EncodeString(name))
	par.Set(root.ParamDescription, codec.EncodeString(description))
	_, err = vmctx.Call(root.Interface.Hname(), coretypes.Hn(root.FuncDeployContract), par, nil)
	return err
}
