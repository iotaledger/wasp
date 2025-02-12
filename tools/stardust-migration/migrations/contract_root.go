package migrations

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/allmigrations"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/newstate"
	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/oldstate"
	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_root "github.com/nnikolash/wasp-types-exported/packages/vm/core/root"
	"github.com/samber/lo"
)

// returns old schema version
func MigrateRootContract(oldChainState old_kv.KVStoreReader, newChainState state.StateDraft) old_isc.SchemaVersion {
	oldContractState := oldstate.GetContactStateReader(oldChainState, old_root.Contract.Hname())
	newContractState := newstate.GetContactState(newChainState, root.Contract.Hname())

	oldSchemaVersion := migrateSchemaVersion(oldChainState, newContractState)
	migrateContractRegistry(oldContractState, newContractState)

	// VarDeployPermissionsEnabled ignored
	// VarDeployPermissions ignored

	return oldSchemaVersion
}

func migrateSchemaVersion(oldChainState old_kv.KVStoreReader, newContractState kv.KVStore) old_isc.SchemaVersion {
	oldRootState := old_root.NewStateAccess(oldChainState)
	oldSchemaVersion := oldRootState.SchemaVersion()
	// oldSchemaVersion is not used anyhow, it is just returned, because new schema is written into new state

	newRootState := root.NewStateWriter(newContractState)
	newRootState.SetSchemaVersion(allmigrations.LatestSchemaVersion)

	return oldSchemaVersion
}

func migrateContractRegistry(oldContractState old_kv.KVStoreReader, newContractState kv.KVStore) {
	newRootState := root.NewStateWriter(newContractState)

	oldRegistry := old_root.GetContractRegistryR(oldContractState)
	newRegistry := newRootState.GetContractRegistry()

	oldRegistry.Iterate(func(key []byte, value []byte) bool {
		oldContractHname := lo.Must(old_isc.HnameFromBytes(key))
		oldContractRecord := lo.Must(old_root.ContractRecordFromBytes(value))

		newContractHname := OldHnameToNewHname(oldContractHname)
		newContractRecord := root.ContractRecord{
			Name: oldContractRecord.Name,
		}

		newRegistry.SetAt(newContractHname.Bytes(), newContractRecord.Bytes())
		return true
	})
}
