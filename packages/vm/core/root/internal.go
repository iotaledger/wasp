package root

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/isc/coreutil"
	"github.com/iotaledger/wasp/v2/packages/kv/collections"
	"github.com/iotaledger/wasp/v2/packages/vm/core/errors/coreerrors"
)

func (s *StateWriter) SetInitialState(v isc.SchemaVersion, contracts []*coreutil.ContractInfo) {
	s.SetSchemaVersion(v)

	contractRegistry := s.GetContractRegistry()
	if contractRegistry.Len() != 0 {
		panic("contract registry must be empty on chain start")
	}

	for _, c := range contracts {
		s.StoreContractRecord(ContractRecordFromContractInfo(c))
	}
}

var errContractAlreadyExists = coreerrors.Register("contract with hname %08x already exists")

func (s *StateWriter) StoreContractRecord(rec *ContractRecord) {
	hname := isc.Hn(rec.Name)
	// storing contract record in the registry
	contractRegistry := s.GetContractRegistry()
	if contractRegistry.HasAt(hname.Bytes()) {
		panic(errContractAlreadyExists.Create(hname))
	}
	contractRegistry.SetAt(hname.Bytes(), rec.Bytes())
}

func (s *StateWriter) GetContractRegistry() *collections.Map {
	return collections.NewMap(s.state, varContractRegistry)
}

func (s *StateReader) GetContractRegistry() *collections.ImmutableMap {
	return collections.NewMapReadOnly(s.state, varContractRegistry)
}

// FindContract is an internal utility function which finds a contract in the KVStore
// It is called from within the 'root' contract as well as VMContext and viewcontext objects
// It is not directly exposed to the sandbox
// If contract is not found by the given hname, nil is returned
func (s *StateReader) FindContract(hname isc.Hname) *ContractRecord {
	contractRegistry := s.GetContractRegistry()
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
