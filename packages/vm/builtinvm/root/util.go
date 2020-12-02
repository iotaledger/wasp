package root

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/coret"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/datatypes"
)

func FindContract(state codec.ImmutableMustCodec, hname coret.Hname) (*ContractRecord, error) {
	if hname == Interface.Hname() {
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

func StoreContract(state codec.ImmutableMustCodec, rec *ContractRecord) error {
	hname := coret.Hn(rec.Name)
	contractRegistry := state.GetMap(VarContractRegistry)
	if contractRegistry.HasAt(hname.Bytes()) {
		return fmt.Errorf("contract with hname %s (name = %s) already exist", hname.String(), rec.Name)
	}
	contractRegistry.SetAt(hname.Bytes(), EncodeContractRecord(rec))
	return nil
}

func DecodeContractRegistry(contractRegistry *datatypes.MustMap) (map[coret.Hname]*ContractRecord, error) {
	ret := make(map[coret.Hname]*ContractRecord)
	var err error
	contractRegistry.Iterate(func(k []byte, v []byte) bool {
		var deploymentHash coret.Hname
		deploymentHash, err = coret.NewHnameFromBytes(k)
		if err != nil {
			return false
		}

		var cr *ContractRecord
		cr, err = DecodeContractRecord(v)
		if err != nil {
			return false
		}

		ret[deploymentHash] = cr
		return true
	})
	return ret, err
}
