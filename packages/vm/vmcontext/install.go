package vmcontext

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
)

// installProgram is a privileged call for root contract
func (vmctx *VMContext) InstallContract(vmtype string, programBinary []byte, name string, description string) error {
	if vmctx.CurrentContractHname() != root.Hname {
		panic("DeployBuiltinContract must be called from root contract")
	}
	vmctx.log.Debugf("VMContext.InstallContract.begin")
	deploymentHash, err := vmctx.processors.NewProcessor(programBinary, vmtype)
	if err != nil {
		return err
	}
	// processor loaded
	vmctx.log.Debugf("VMContext.InstallContract.1")

	state := codec.NewMustCodec(vmctx)

	// if program binary is not in the registry, write it there
	binRegistry := state.GetMap(root.VarRegistryOfBinaries)
	if !binRegistry.HasAt(deploymentHash[:]) {
		binRegistry.SetAt(deploymentHash[:], programBinary)
	}
	hname := coretypes.Hn(name)
	contractRegistry := state.GetMap(root.VarContractRegistry)
	if contractRegistry.HasAt(hname.Bytes()) {
		return fmt.Errorf("contract with hname %s (name = %s) already exist", hname.String(), name)
	}
	contractRegistry.SetAt(hname.Bytes(), root.EncodeContractRecord(&root.ContractRecord{
		VMType:         vmtype,
		DeploymentHash: *deploymentHash,
		Description:    description,
		Name:           name,
	}))

	return nil
}
