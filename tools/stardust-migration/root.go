package main

import (
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_root "github.com/nnikolash/wasp-types-exported/packages/vm/core/root"
)

func migrateRootContract(srcChainState old_kv.KVStoreReader, destChainState state.StateDraft) {
	migrateRecord(srcChainState, destChainState, old_root.VarSchemaVersion, asIs[uint32](root.VarSchemaVersion))

	migrateRecord(srcChainState, destChainState, old_root.VarContractRegistry,
		func(k old_kv.Key, r old_root.ContractRecord) (kv.Key, root.ContractRecord) {
			return kv.Key(k), root.ContractRecord{
				Name: r.Name,
			}
		})

	// VarDeployPermissionsEnabled ignored
	// VarDeployPermissions ignored
}
