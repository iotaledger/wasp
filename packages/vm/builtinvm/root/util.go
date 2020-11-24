package root

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

func FindContract(state codec.ImmutableMustCodec, hname coretypes.Hname) (*ContractRecord, error) {
	if hname == Hname {
		return &RootContractRecord, nil
	}
	contractRegistry := state.GetMap(VarContractRegistry)
	retBin := contractRegistry.GetAt(hname.Bytes())
	if retBin == nil {
		return nil, fmt.Errorf("root: contract %s not found", hname)
	}
	ret, err := DecodeContractRecord(retBin)
	if err != nil {
		return nil, fmt.Errorf("root: %v", err)
	}
	return ret, nil
}

func GetBinary(state codec.ImmutableMustCodec, deploymentHash hashing.HashValue) ([]byte, error) {
	contractRegistry := state.GetMap(VarRegistryOfBinaries)
	if ret := contractRegistry.GetAt(deploymentHash[:]); ret != nil {
		return ret, nil
	}
	return nil, fmt.Errorf("binary not found")
}

func StoreContract(state codec.ImmutableMustCodec, rec *ContractRecord, programBinary []byte) error {
	binRegistry := state.GetMap(VarRegistryOfBinaries)
	if !binRegistry.HasAt(rec.DeploymentHash[:]) {
		binRegistry.SetAt(rec.DeploymentHash[:], programBinary)
	}
	hname := coretypes.Hn(rec.Name)
	contractRegistry := state.GetMap(VarContractRegistry)
	if contractRegistry.HasAt(hname.Bytes()) {
		return fmt.Errorf("contract with hname %s (name = %s) already exist", hname.String(), rec.Name)
	}
	contractRegistry.SetAt(hname.Bytes(), EncodeContractRecord(rec))
	return nil
}
