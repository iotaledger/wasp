package migrations

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
)

type Migration struct {
	Contract *coreutil.ContractInfo
	Apply    func(contractState kv.KVStore, log *logger.Logger) error
}

type MigrationScheme struct {
	BaseSchemaVersion uint32
	Migrations        []Migration
}

func (m *MigrationScheme) LatestSchemaVersion() uint32 {
	return m.BaseSchemaVersion + uint32(len(m.Migrations))
}
