package vmcontext

import (
	"github.com/iotaledger/wasp/packages/coret"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
)

// CreateContract deploys contract by its program hash
func (vmctx *VMContext) CreateContract(programHash hashing.HashValue, name string, description string, initParams codec.ImmutableCodec) error {
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
	par := codec.NewCodec(dict.New())
	err = initParams.Iterate("", func(key kv.Key, value []byte) bool {
		par.Set(key, value)
		return true
	})
	if err != nil {
		return err
	}
	par.SetHashValue(root.ParamProgramHash, &programHash)
	par.SetString(root.ParamName, name)
	par.SetString(root.ParamDescription, description)
	_, err = vmctx.Call(root.Interface.Hname(), coret.Hn(root.FuncDeployContract), par, nil)
	return err

}
