package root

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"golang.org/x/xerrors"
)

func GetContractRegistry(state kv.KVStore) *collections.Map {
	return collections.NewMap(state, StateVarContractRegistry)
}

func GetContractRegistryR(state kv.KVStoreReader) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, StateVarContractRegistry)
}

// FindContract is an internal utility function which finds a contract in the KVStore
// It is called from within the 'root' contract as well as VMContext and viewcontext objects
// It is not directly exposed to the sandbox
// If contract is not found by the given hname, nil is returned
func FindContract(state kv.KVStoreReader, hname iscp.Hname) *ContractRecord {
	contractRegistry := GetContractRegistryR(state)
	retBin := contractRegistry.MustGetAt(hname.Bytes())
	if retBin != nil {
		ret, err := ContractRecordFromBytes(retBin)
		if err != nil {
			panic(xerrors.Errorf("FindContract: %w", err))
		}
		return ret
	}
	if hname == Contract.Hname() {
		// it happens during bootstrap
		return NewContractRecord(Contract, &iscp.NilAgentID)
	}
	return nil
}

func ContractExists(state kv.KVStoreReader, hname iscp.Hname) bool {
	return GetContractRegistryR(state).MustHasAt(hname.Bytes())
}

// DecodeContractRegistry encodes the whole contract registry from the map into a Go map.
func DecodeContractRegistry(contractRegistry *collections.ImmutableMap) (map[iscp.Hname]*ContractRecord, error) {
	ret := make(map[iscp.Hname]*ContractRecord)
	var err error
	contractRegistry.MustIterate(func(k []byte, v []byte) bool {
		var deploymentHash iscp.Hname
		deploymentHash, err = iscp.HnameFromBytes(k)
		if err != nil {
			return false
		}

		cr, err := ContractRecordFromBytes(v)
		if err != nil {
			return false
		}

		ret[deploymentHash] = cr
		return true
	})
	return ret, err
}
