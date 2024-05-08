package root

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/collections"
)

func GetContractRegistry(state kv.KVStore) *collections.Map {
	return collections.NewMap(state, VarContractRegistry)
}

func GetContractRegistryR(state kv.KVStoreReader) *collections.ImmutableMap {
	return collections.NewMapReadOnly(state, VarContractRegistry)
}

// FindContract is an internal utility function which finds a contract in the KVStore
// It is called from within the 'root' contract as well as VMContext and viewcontext objects
// It is not directly exposed to the sandbox
// If contract is not found by the given hname, nil is returned
func FindContract(state kv.KVStoreReader, hname isc.Hname) *ContractRecord {
	contractRegistry := GetContractRegistryR(state)
	retBin := contractRegistry.GetAt(hname.Bytes())
	if retBin != nil {
		ret, err := ContractRecordFromBytes(retBin)
		if err != nil {
			panic(fmt.Errorf("FindContract: %w", err))
		}
		return ret
	}
	if hname == Contract.Hname() {
		// it happens during bootstrap
		return ContractRecordFromContractInfo(Contract)
	}
	return nil
}

// decodeContractRegistry encodes the whole contract registry from the map into a Go map.
func decodeContractRegistry(contractRegistry *collections.ImmutableMap) (map[isc.Hname]*ContractRecord, error) {
	ret := make(map[isc.Hname]*ContractRecord)
	var err error
	contractRegistry.Iterate(func(k []byte, v []byte) bool {
		var deploymentHash isc.Hname
		deploymentHash, err = isc.HnameFromBytes(k)
		if err != nil {
			return false
		}

		cr, err2 := ContractRecordFromBytes(v)
		if err2 != nil {
			return false
		}

		ret[deploymentHash] = cr
		return true
	})
	return ret, err
}
