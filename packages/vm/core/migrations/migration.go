package migrations

import (
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/log"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
)

type Migration struct {
	Contract *coreutil.ContractInfo
	Apply    func(contractState kv.KVStore, log log.Logger) error
}

type MigrationScheme struct {
	BaseSchemaVersion isc.SchemaVersion
	Migrations        []Migration
}

func (m *MigrationScheme) LatestSchemaVersion() isc.SchemaVersion {
	return m.BaseSchemaVersion + isc.SchemaVersion(len(m.Migrations))
}

var (
	ErrMissingMigrationCode = errors.New("missing migration code for target schema version")
	ErrInvalidSchemaVersion = errors.New("invalid schema version")
)

// WithTargetSchemaVersion returns a new MigrationScheme where all migrations
// that correspond to a schema version newer than v are removed.
// This is necessary in order to replay old blocks without applying the newer migrations.
func (m *MigrationScheme) WithTargetSchemaVersion(v isc.SchemaVersion) (*MigrationScheme, error) {
	newMigrations := m.Migrations
	if len(newMigrations) > 0 {
		if v < m.BaseSchemaVersion {
			return nil, fmt.Errorf("cannot determine migration scheme for target schema version %d: %w", v, ErrMissingMigrationCode)
		}
		if v > m.LatestSchemaVersion() {
			return nil, fmt.Errorf("cannot determine migration scheme for target schema version %d: %w", v, ErrInvalidSchemaVersion)
		}
		newMigrations = newMigrations[:v-m.BaseSchemaVersion]
	}
	return &MigrationScheme{
		BaseSchemaVersion: m.BaseSchemaVersion,
		Migrations:        newMigrations,
	}, nil
}
