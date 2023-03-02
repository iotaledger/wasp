package migrations

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
)

const BaseSchemaVersion = uint32(0)

type Migration struct {
	Contract *coreutil.ContractInfo
	Apply    func(state kv.KVStore, log *logger.Logger) error
}

// Add new migrations to the end of this list, and they will be applied before
// creating the next block.
// The first migration on the list is applied when schema version =
// BaseSchemaVersion, and after applying each migration the schema version is
// incremented.
// Old migrations can be pruned; for each migration pruned increment
// BaseSchemaVersion by one.
var Migrations = []Migration{
	m001GasPerTokenToRatio32,
	m002CleanupFeePolicy,
}
