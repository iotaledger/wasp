package vmcontext

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/core/migrations"
	"github.com/iotaledger/wasp/packages/vm/core/root"
)

func (vmctx *VMContext) runMigrations() {
	if vmctx.task.AnchorOutput.StateIndex == 0 {
		// initializing new chain -- set the schema to latest version
		vmctx.callCore(root.Contract, func(s kv.KVStore) {
			root.SetSchemaVersion(s, migrations.LatestSchemaVersion)
		})
		return
	}

	var schemaVersion uint32
	vmctx.callCore(root.Contract, func(s kv.KVStore) {
		schemaVersion = root.GetSchemaVersion(s)
	})
	if schemaVersion < migrations.BaseSchemaVersion {
		panic(fmt.Sprintf("inconsistency: schema version %d should be >= %d", schemaVersion, migrations.BaseSchemaVersion))
	}
	if schemaVersion > migrations.LatestSchemaVersion {
		panic(fmt.Sprintf("inconsistency: schema version %d should be <= %d", schemaVersion, migrations.LatestSchemaVersion))
	}

	for schemaVersion < migrations.LatestSchemaVersion {
		migration := migrations.Migrations[schemaVersion-migrations.BaseSchemaVersion]

		vmctx.callCore(migration.Contract, func(s kv.KVStore) {
			err := migration.Apply(s, vmctx.task.Log)
			if err != nil {
				panic(fmt.Sprintf("failed applying migration: %s", err))
			}
		})

		schemaVersion++
		vmctx.callCore(root.Contract, func(s kv.KVStore) {
			root.SetSchemaVersion(s, schemaVersion)
		})
	}
}
