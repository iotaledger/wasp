package migrations

import (
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_root "github.com/nnikolash/wasp-types-exported/packages/vm/core/root"
)

func MigrateRootContract(oldChainState old_kv.KVStoreReader, newChainState state.StateDraft) {
	MigrateVariable(oldChainState, newChainState, old_root.VarSchemaVersion, root.VarSchemaVersion, AsIs[uint32])

	MigrateVariable(oldChainState, newChainState, old_root.VarContractRegistry, root.VarContractRegistry,
		func(r old_root.ContractRecord) root.ContractRecord {
			return root.ContractRecord{
				Name: r.Name,
			}
		})

	// VarDeployPermissionsEnabled ignored
	// VarDeployPermissions ignored
}
