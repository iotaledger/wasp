// Package allmigrations defines all migrations to be applied with rebased
package allmigrations

import (
	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/vm/core/migrations"
	"github.com/iotaledger/wasp/v2/packages/vm/core/root"
)

const (
	// versions prior to 5 correspond to stardust
	// version 5 acts as a marker for migrated stardust blocks in case legacy behavior needs to be introduced.
	SchemaVersionMigratedRebased = 5 + iota
	SchemaVersionIotaRebased

	LatestSchemaVersion = SchemaVersionIotaRebased
)

var DefaultScheme = &migrations.MigrationScheme{
	BaseSchemaVersion: SchemaVersionMigratedRebased,

	// Add new migrations to the end of this list, and they will be applied before
	// creating the next block.
	// The first migration on the list is applied when schema version =
	// BaseSchemaVersion, and after applying each migration the schema version is
	// incremented.
	// Old migrations can be pruned; for each migration pruned increment
	// BaseSchemaVersion by one.
	Migrations: []migrations.Migration{
		// This adds a NOOP migration, enabling the proper handling of migrated blocks (legacy encoding, mostly)
		// and making sure that new blocks are created with SchemaVersionIotaRebased.
		{
			Apply: func(contractState kv.KVStore, log log.Logger) error {
				return nil
			},
			Contract: root.Contract,
		},
	},
}
