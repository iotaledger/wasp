package migrations

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
)

type Migration struct {
	Contract *coreutil.ContractInfo
	Apply    func(contractState kv.KVStore, log *logger.Logger) error
}

type MigrationScheme struct {
	BaseSchemaVersion isc.SchemaVersion
	Migrations        []Migration
}

func (m *MigrationScheme) LatestSchemaVersion() isc.SchemaVersion {
	return m.BaseSchemaVersion + isc.SchemaVersion(len(m.Migrations))
}
