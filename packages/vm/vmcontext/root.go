package vmcontext

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
)

// installProgram is a privileged call for root contract
func (vmctx *VMContext) InstallContract(vmtype string, programBinary []byte, name string, description string) error {
	if vmctx.ContractHname() != root.ContractHname {
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

func (vmctx *VMContext) findContractByHname(contractHname coretypes.Hname) (*root.ContractRecord, bool) {
	if contractHname == root.ContractHname {
		// root
		return root.GetRootContractRecord(), true
	}
	params := codec.NewCodec(dict.NewDict())
	params.SetInt64(root.ParamHname, int64(contractHname))
	res, err := vmctx.callRoot(root.EntryPointFindContract, params)
	if err != nil {
		return nil, false
	}
	data, err := res.Get(root.ParamData)
	if err != nil {
		return nil, false
	}
	ret, err := root.DecodeContractRecord(data)
	if err != nil {
		return nil, false
	}
	return ret, true
}

func (vmctx *VMContext) getBinary(deploymentHash *hashing.HashValue) ([]byte, bool) {
	params := codec.NewCodec(dict.NewDict())
	params.SetHashValue(root.ParamHash, deploymentHash)
	res, err := vmctx.callRoot(root.EntryPointGetBinary, params)
	if err != nil {
		return nil, false
	}
	data, err := res.Get(root.ParamData)
	if err != nil {
		return nil, false
	}
	return data, true
}

func (vmctx *VMContext) getProcessor(rec *root.ContractRecord) (vmtypes.Processor, error) {
	if proc, ok := vmctx.processors.GetProcessor(&rec.DeploymentHash); ok {
		return proc, nil
	}
	// load processor to the cache
	binary, ok := vmctx.getBinary(&rec.DeploymentHash)
	if !ok {
		return nil, fmt.Errorf("internal error: can't get the binary for the program")
	}
	deploymentHash, err := vmctx.processors.NewProcessor(binary, rec.VMType)
	if err != nil {
		return nil, err
	}
	if *deploymentHash != rec.DeploymentHash {
		return nil, fmt.Errorf("internal error: *deploymentHash != rec.DeploymentHash")
	}
	if proc, ok := vmctx.processors.GetProcessor(deploymentHash); ok {
		return proc, nil
	}
	return nil, fmt.Errorf("internal error: can't get the deployed processor")
}

func (vmctx *VMContext) callRoot(entryPointCode coretypes.Hname, params codec.ImmutableCodec) (codec.ImmutableCodec, error) {
	return vmctx.CallContract(root.ContractHname, entryPointCode, params, nil)
}
