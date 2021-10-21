package root

import (
	"golang.org/x/xerrors"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/vm/core/_default"
)

// FindContract is an internal utility function which finds a contract in the KVStore
// It is called from within the 'root' contract as well as VMContext and viewcontext objects
// It is not directly exposed to the sandbox
// If contract is not found by the given hname, the default contract is returned
// the bool flag indicates if contract was found or not. It means, if hname == 0 it will
// return default contract and true
func FindContract(state kv.KVStoreReader, hname iscp.Hname) (*ContractRecord, bool) {
	contractRegistry := collections.NewMapReadOnly(state, VarContractRegistry)
	retBin := contractRegistry.MustGetAt(hname.Bytes())
	if retBin != nil {
		ret, err := ContractRecordFromBytes(retBin)
		if err != nil {
			panic(xerrors.Errorf("FindContract: %w", err))
		}
		return ret, true
	}
	if hname == Contract.Hname() {
		// it happens during bootstrap
		return NewContractRecord(Contract, &iscp.NilAgentID), true
	}
	return NewContractRecord(_default.Contract, &iscp.NilAgentID), false
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
