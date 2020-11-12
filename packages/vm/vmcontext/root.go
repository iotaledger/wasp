package vmcontext

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
)

// installProgram is a privileged call for root contract
func (vmctx *VMContext) InstallContract(vmtype string, programBinary []byte, name string, description string) (uint16, error) {
	if vmctx.ContractIndex() != 0 {
		panic("DeployBuiltinContract must be called from root contract")
	}
	vmctx.log.Debugf("VMContext.InstallContract.begin")
	deploymentHash, err := vmctx.processors.NewProcessor(programBinary, vmtype)
	if err != nil {
		return 0, err
	}
	// processor loaded
	vmctx.log.Debugf("VMContext.InstallContract.1")

	state := codec.NewMustCodec(vmctx)

	// if program binary is not in the registry, write it there
	binRegistry := state.GetMap(root.VarRegistryOfBinaries)
	if !binRegistry.HasAt(deploymentHash[:]) {
		binRegistry.SetAt(deploymentHash[:], programBinary)
	}

	contractRegistry := state.GetArray(root.VarContractRegistry)
	contractRegistry.Push(root.EncodeContractRecord(&root.ContractRecord{
		VMType:         vmtype,
		DeploymentHash: *deploymentHash,
		Description:    description,
		Name:           name,
	}))

	return contractRegistry.Len() - 1, nil
}

func (vmctx *VMContext) getBinary(deploymentHash *hashing.HashValue) ([]byte, bool) {
	return root.GetBinary(deploymentHash, vmctx.callRoot)
}

func (vmctx *VMContext) callRoot(entryPointCode coretypes.Hname, params codec.ImmutableCodec) (codec.ImmutableCodec, error) {
	return vmctx.CallContract(0, entryPointCode, params, nil)
}
