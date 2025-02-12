package migrations

import (
	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_root "github.com/nnikolash/wasp-types-exported/packages/vm/core/root"

	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/allmigrations"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/newstate"
	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/oldstate"
)

// returns old schema version
func MigrateRootContract(oldChainState old_kv.KVStoreReader, newChainState state.StateDraft) old_isc.SchemaVersion {
	oldContractState := oldstate.GetContactStateReader(oldChainState, old_root.Contract.Hname())
	newContractState := newstate.GetContactState(newChainState, root.Contract.Hname())

	v, _ := MigrateVariable(oldContractState, newContractState, old_root.VarSchemaVersion, root.VarSchemaVersion,
		func(uint32) uint32 {
			return allmigrations.LatestSchemaVersion
		})

	MigrateVariable(oldContractState, newContractState, old_root.VarContractRegistry, root.VarContractRegistry,
		func(r old_root.ContractRecord) root.ContractRecord {
			return root.ContractRecord{
				Name: r.Name,
			}
		})

	// VarDeployPermissionsEnabled ignored
	// VarDeployPermissions ignored

	return old_isc.SchemaVersion(v)
}
