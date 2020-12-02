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
	vmctx.log.Debugf("vmcontext.DeployContract: %s, name: %s, dscr: '%s'", programHash.String(), name, description)

	vmtype, programBinary, err := vmctx.getBinary(programHash)
	if err != nil {
		return err
	}
	if vmctx.CurrentContractHname() == root.Interface.Hname() {
		// from root contract calling VMContext directly
		err := vmctx.processors.NewProcessor(programHash, programBinary, vmtype)
		if err != nil {
			return err
		}
		// storing contract in the registry
		err = root.StoreContract(codec.NewMustCodec(vmctx), &root.ContractRecord{
			ProgramHash: programHash,
			Description: description,
			Name:        name,
		})
		if err != nil {
			return err
		}
		// calling constructor
		_, err = vmctx.Call(coret.Hn(name), coret.EntryPointInit, initParams, nil)
		if err != nil {
			vmctx.log.Warnf("sandbox.CreateContract: calling init function: %v", err)
		}
		// ignoring error because init method may not exist
		return nil
	}
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
