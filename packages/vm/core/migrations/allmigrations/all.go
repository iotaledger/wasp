package allmigrations

import (
	"github.com/iotaledger/wasp/packages/vm/core/migrations"
)

const (
	SchemaVersionMigratedRebased = 5 + iota
	SchemaVersionIotaRebased     // versions prior to 4 correspond to stardust

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
	Migrations: []migrations.Migration{},
}
