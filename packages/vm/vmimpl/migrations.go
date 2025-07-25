package vmimpl

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/vm/core/migrations"
	"github.com/iotaledger/wasp/v2/packages/vm/core/root"
)

func (vmctx *vmContext) runMigrations(chainState kv.KVStore, migrationScheme *migrations.MigrationScheme) {
	latestSchemaVersion := migrationScheme.LatestSchemaVersion()

	currentVersion := root.NewStateReaderFromChainState(chainState).GetSchemaVersion()

	if currentVersion < migrationScheme.BaseSchemaVersion {
		panic(fmt.Sprintf("inconsistency: node with schema version %d is behind pruned migrations (should be >= %d)", currentVersion, migrationScheme.BaseSchemaVersion))
	}
	if currentVersion > latestSchemaVersion {
		panic(fmt.Sprintf("inconsistency: node with schema version %d is ahead latest schema version (should be <= %d)", currentVersion, latestSchemaVersion))
	}

	for currentVersion < latestSchemaVersion {
		migration := migrationScheme.Migrations[currentVersion-migrationScheme.BaseSchemaVersion]

		err := migration.Apply(migration.Contract.StateSubrealm(chainState), vmctx.task.Log)
		if err != nil {
			panic(fmt.Sprintf("failed applying migration: %s", err))
		}

		currentVersion++
		root.NewStateWriter(root.Contract.StateSubrealm(chainState)).SetSchemaVersion(currentVersion)
	}
}
