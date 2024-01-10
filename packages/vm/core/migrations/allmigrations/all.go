package allmigrations

import (
	"github.com/iotaledger/wasp/packages/vm/core/migrations"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/m001"
)

var DefaultScheme = &migrations.MigrationScheme{
	BaseSchemaVersion: 0,

	// Add new migrations to the end of this list, and they will be applied before
	// creating the next block.
	// The first migration on the list is applied when schema version =
	// BaseSchemaVersion, and after applying each migration the schema version is
	// incremented.
	// Old migrations can be pruned; for each migration pruned increment
	// BaseSchemaVersion by one.
	Migrations: []migrations.Migration{
		m001.AccountDecimals,
	},
}
